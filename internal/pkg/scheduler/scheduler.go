package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	constraint_policy_client "github.com/ciena/turnbuckle/internal/pkg/constraint-policy-client"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// ConstraintPolicyScheduler defines the runtime information used by the
// constraint policy scheduler.
type ConstraintPolicyScheduler struct {
	options               ConstraintPolicySchedulerOptions
	log                   logr.Logger
	defaultPlanner        *ConstraintPolicySchedulerPlanner
	fh                    framework.Handle
	quit                  chan struct{}
	podQueue              chan *v1.Pod
	podRequeueQueue       workqueue.RateLimitingInterface
	podRequeueMap         map[ObjectMeta]struct{}
	constraintPolicyMutex sync.Mutex
	podRequeueMutex       sync.Mutex
}

// ConstraintPolicySchedulerOptions defines the configuration options for the
// constraint policy scheduler.
type ConstraintPolicySchedulerOptions struct {
	Planner              bool
	NumRetriesOnFailure  int
	MinDelayOnFailure    time.Duration
	MaxDelayOnFailure    time.Duration
	FallbackOnNoOffers   bool
	RetryOnNoOffers      bool
	RequeuePeriod        time.Duration
	PodQueueSize         uint
	PlannerNodeQueueSize uint
}

// NewScheduler create a new instance of the schedule logic with the
// given set or parameters.
func NewScheduler(options ConstraintPolicySchedulerOptions,
	clientset kubernetes.Interface,
	fh framework.Handle,
	constraintPolicyClient constraint_policy_client.ConstraintPolicyClient,
	log logr.Logger) *ConstraintPolicyScheduler {
	var addPodCallback, deletePodCallback func(pod *v1.Pod)

	constraintPolicyScheduler := &ConstraintPolicyScheduler{}
	podQueue := make(chan *v1.Pod, options.PodQueueSize)

	deletePodCallback = func(pod *v1.Pod) {
		constraintPolicyScheduler.handlePodDelete(pod)
	}

	defaultPlanner := NewPlanner(
		ConstraintPolicySchedulerPlannerOptions{
			NodeQueueSize:     options.PlannerNodeQueueSize,
			AddPodCallback:    addPodCallback,
			DeletePodCallback: deletePodCallback,
		},
		clientset, constraintPolicyClient, log.WithName("default-planner"))

	constraintPolicyScheduler.options = options
	constraintPolicyScheduler.fh = fh
	constraintPolicyScheduler.defaultPlanner = defaultPlanner
	constraintPolicyScheduler.log = log
	constraintPolicyScheduler.podQueue = podQueue
	constraintPolicyScheduler.quit = make(chan struct{})
	constraintPolicyScheduler.podRequeueQueue = workqueue.
		NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(
			options.MinDelayOnFailure,
			options.MaxDelayOnFailure))
	constraintPolicyScheduler.podRequeueMap = make(map[ObjectMeta]struct{})

	return constraintPolicyScheduler
}

func (s *ConstraintPolicyScheduler) handlePodDelete(pod *v1.Pod) {
	s.podRequeueMutex.Lock()
	defer s.podRequeueMutex.Unlock()
	delete(s.podRequeueMap, ObjectMeta{Name: pod.Name, Namespace: pod.Namespace})
}

// Stop halts the constraint policy scheduler.
func (s *ConstraintPolicyScheduler) Stop() {
	s.podRequeueQueue.ShutDown()
	close(s.quit)
}

// Start invokes the constraint policy scheduler as a go routine.
func (s *ConstraintPolicyScheduler) Start() {
	go s.listenForPodRequeueEvents()
}

func (s *ConstraintPolicyScheduler) processRequeueEvents() bool {
	// get blocks till there is an item
	item, quit := s.podRequeueQueue.Get()
	if quit {
		return false
	}

	defer s.podRequeueQueue.Done(item)

	s.processRequeue(item)

	return true
}

func (s *ConstraintPolicyScheduler) processRequeue(item interface{}) bool {
	data, ok := item.(*v1.Pod)
	if !ok {
		s.log.V(1).Info("pod-requeue-item-not-a-pod")
		s.podRequeueQueue.Forget(item)

		return false
	}
	// if the item is not in the requeue map, forget it
	s.podRequeueMutex.Lock()
	defer s.podRequeueMutex.Unlock()

	forgetItem := true

	defer func() {
		if forgetItem {
			s.podRequeueQueue.Forget(item)
			delete(s.podRequeueMap, ObjectMeta{Name: data.Name, Namespace: data.Namespace})
		}
	}()

	pod, err := s.defaultPlanner.
		GetClientset().CoreV1().Pods(data.Namespace).
		Get(context.Background(), data.Name, metav1.GetOptions{})
	if err != nil {
		s.log.Error(err, "pod-requeue-pod-get-failure", "pod", data.Name)

		return false
	}

	if !s.podCheckScheduler(pod) {
		s.log.V(1).Info("pod-requeue-scheduler-mismatch", "pod", pod.Name)

		return false
	}

	if pod.Spec.NodeName != "" {
		s.log.V(1).Info("pod-requeue-node-assigned", "pod", pod.Name, "node", pod.Spec.NodeName)

		return false
	}

	if pod.Status.Phase != v1.PodPending {
		s.log.V(1).Info("pod-requeue-status-not-pending", "pod", pod.Name)

		return false
	}

	if _, ok := s.podRequeueMap[ObjectMeta{Name: pod.Name, Namespace: pod.Namespace}]; !ok {
		s.log.V(1).Info("pod-requeue-delete", "pod-probably-deleted", pod.Name)

		return false
	}

	numRequeues := s.podRequeueQueue.NumRequeues(item)
	if s.options.NumRetriesOnFailure <= 0 || numRequeues <= s.options.NumRetriesOnFailure {
		forgetItem = false
		// add back the pod to the podqueue
		s.log.V(1).Info("pod-requeue-add", "pod", pod.Name, "num-requeues", numRequeues)

		return true
	}

	s.log.V(1).Info("pod-requeue-requeues-exceeded", "pod", pod.Name)

	return false
}

func (s *ConstraintPolicyScheduler) listenForPodRequeueEvents() {
	defer s.podRequeueQueue.ShutDown()

	go wait.Until(s.requeueWorker, s.options.RequeuePeriod, s.quit)

	<-s.quit
}

func (s *ConstraintPolicyScheduler) requeueWorker() {
	for s.processRequeueEvents() {
	}
}

// requeue pod into the worker queue.
func (s *ConstraintPolicyScheduler) requeue(pod *v1.Pod) bool {
	s.podRequeueMutex.Lock()

	if _, ok := s.podRequeueMap[ObjectMeta{Name: pod.Name, Namespace: pod.Namespace}]; !ok {
		s.podRequeueMap[ObjectMeta{Name: pod.Name, Namespace: pod.Namespace}] = struct{}{}
	}

	s.log.V(1).Info("pod-requeue", "pod", pod.Name, "namespace", pod.Namespace)
	s.podRequeueQueue.AddRateLimited(pod)
	s.podRequeueMutex.Unlock()

	// retry from a rate limited queue
	// blocks till item is available
	item, quit := s.podRequeueQueue.Get()
	if quit {
		return false
	}

	defer s.podRequeueQueue.Done(item)

	return s.processRequeue(item)
}

func (s *ConstraintPolicyScheduler) requeueFixup(pod *v1.Pod) {
	s.podRequeueMutex.Lock()
	defer s.podRequeueMutex.Unlock()
	delete(s.podRequeueMap, ObjectMeta{Name: pod.Name, Namespace: pod.Namespace})
	s.podRequeueQueue.Forget(pod)
}

func (s *ConstraintPolicyScheduler) podCheckScheduler(pod *v1.Pod) bool {
	return true
}

func (s *ConstraintPolicyScheduler) findFit(pod *v1.Pod, eligibleNodes []*v1.Node) (*v1.Node, error) {
	var nodeInstance *v1.Node

	var err error

	for {
		nodeInstance, err = s.defaultPlanner.FindBestNode(pod, eligibleNodes)
		if errors.Is(err, ErrNoOffers) {
			s.log.V(1).Info("no-offers-found-for-pod", "pod", pod.Name)

			if s.options.FallbackOnNoOffers {
				s.log.V(1).Info("random-assignment", "pod", pod.Name)

				return s.defaultPlanner.FindFitRandom(pod, eligibleNodes)
			}

			if s.options.RetryOnNoOffers {
				s.log.V(1).Info("requeuing-for-retry", "pod", pod.Name)

				if !s.requeue(pod) {
					return nil, err
				}

				continue
			}

			return nil, err
		}

		if err != nil {
			s.log.Error(err, "nodes-not-found", "pod", pod.Name)
			s.log.V(1).Info("requeuing-for-retry", "pod", pod.Name)

			if !s.requeue(pod) {
				return nil, err
			}

			continue
		}

		break
	}
	// if the pod was requeued, then cleanup the entry from requeue map
	s.requeueFixup(pod)
	s.log.V(1).Info("found-matching", "node", nodeInstance.Name, "pod", pod.Name)

	return nodeInstance, nil
}

// FindBestNode finds the best node for the pod.
func (s *ConstraintPolicyScheduler) FindBestNode(pod *v1.Pod, feasibleNodes []*v1.Node) (*v1.Node, error) {
	s.log.V(1).Info("find-best-node", "pod", pod.Name)

	s.constraintPolicyMutex.Lock()
	defer s.constraintPolicyMutex.Unlock()

	node, err := s.findFit(pod, feasibleNodes)

	if errors.Is(err, ErrNoOffers) {
		// if no offers are matched for the pod, return the existing feasible nodes
		if len(feasibleNodes) > 0 {
			return feasibleNodes[0], nil
		}

		return nil, err
	}

	if err != nil {
		s.log.V(1).Info("scheduler-plugin", "no-nodes-available-to-schedule-pod", pod.Name)

		return nil, err
	}

	if node == nil {
		s.log.V(1).Info("scheduler-plugin", "pod-waiting-for-planner-assignment", pod.Name)

		return nil, nil
	}

	return node, nil
}

func (s *ConstraintPolicyScheduler) emitEvent(p *v1.Pod, message string) error {
	timestamp := time.Now().UTC()

	_, err := s.defaultPlanner.GetClientset().CoreV1().Events(p.Namespace).Create(context.Background(),
		&v1.Event{
			Count:          1,
			Message:        message,
			Reason:         "Scheduled",
			LastTimestamp:  metav1.NewTime(timestamp),
			FirstTimestamp: metav1.NewTime(timestamp),
			Type:           "Normal",
			Source: v1.EventSource{
				Component: Name,
			},
			InvolvedObject: v1.ObjectReference{
				Kind:      "Pod",
				Name:      p.Name,
				Namespace: p.Namespace,
				UID:       p.UID,
			},
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: p.Name + "-",
			},
		},
		metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error creating event: %w", err)
	}

	return nil
}
