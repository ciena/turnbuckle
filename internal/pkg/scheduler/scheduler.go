package scheduler

import (
	"fmt"
	constraint_policy_client "github.com/ciena/turnbuckle/internal/pkg/constraint-policy-client"
	"github.com/ciena/turnbuckle/internal/pkg/nsm"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	"math/rand"
	"sync"
	"time"
)

const (
	SchedulerName = "constraint-policy-scheduler"
)

type ConstraintPolicyScheduler struct {
	options               ConstraintPolicySchedulerOptions
	log                   logr.Logger
	defaultPlanner        *ConstraintPolicySchedulerPlanner
	quit                  chan struct{}
	podQueue              chan *v1.Pod
	podRequeueQueue       workqueue.RateLimitingInterface
	podRequeueMap         map[ObjectMeta]struct{}
	constraintPolicyMutex sync.Mutex
	podRequeueMutex       sync.Mutex
}

type ConstraintPolicySchedulerOptions struct {
	Extender              bool
	Planner               bool
	NumRetriesOnFailure   int
	MinDelayOnFailure     time.Duration
	MaxDelayOnFailure     time.Duration
	FallbackOnNoOffers    bool
	RetryOnNoOffers       bool
	CheckForDuplicatePods bool
}

func NewScheduler(options ConstraintPolicySchedulerOptions, clientset *kubernetes.Clientset,
	constraintPolicyClient constraint_policy_client.ConstraintPolicyClient,
	networkServiceClient nsm.NetworkServiceClient,
	log logr.Logger) *ConstraintPolicyScheduler {
	var addPodCallback, deletePodCallback func(pod *v1.Pod)
	constraintPolicyScheduler := &ConstraintPolicyScheduler{}
	podQueue := make(chan *v1.Pod, 300)
	addPodCallback = func(pod *v1.Pod) {
		if !options.Extender && pod.Spec.NodeName == "" && pod.Spec.SchedulerName == SchedulerName {
			podQueue <- pod
		}
	}
	deletePodCallback = func(pod *v1.Pod) {
		constraintPolicyScheduler.handlePodDelete(pod)
	}
	defaultPlanner := NewPlannerWithPodCallbacks(ConstraintPolicySchedulerPlannerOptions{
		Extender:              options.Extender,
		CheckForDuplicatePods: options.CheckForDuplicatePods,
	},
		clientset, constraintPolicyClient, networkServiceClient, log.WithName("default-planner"),
		addPodCallback, nil, deletePodCallback)

	constraintPolicyScheduler.options = options
	constraintPolicyScheduler.defaultPlanner = defaultPlanner
	constraintPolicyScheduler.log = log
	constraintPolicyScheduler.podQueue = podQueue
	constraintPolicyScheduler.quit = make(chan struct{})
	constraintPolicyScheduler.podRequeueQueue = workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(options.MinDelayOnFailure,
		options.MaxDelayOnFailure))
	constraintPolicyScheduler.podRequeueMap = make(map[ObjectMeta]struct{})
	return constraintPolicyScheduler
}

func (s *ConstraintPolicyScheduler) handlePodDelete(pod *v1.Pod) {
	s.podRequeueMutex.Lock()
	defer s.podRequeueMutex.Unlock()
	delete(s.podRequeueMap, ObjectMeta{Name: pod.Name, Namespace: pod.Namespace})
}

func (s *ConstraintPolicyScheduler) Stop() {
	if s.options.Extender {
		s.podRequeueQueue.ShutDown()
	}
	close(s.quit)
}

// to be used with scheduler extender option
func (s *ConstraintPolicyScheduler) Start() {
	go s.listenForPodRequeueEvents()
}

func (s *ConstraintPolicyScheduler) Run() error {
	go s.listenForPodRequeueEvents()
	for {
		select {
		case p := <-s.podQueue:
			s.constraintPolicyMutex.Lock()
			node, err := s.findFit(p, []v1.Node{})
			if err != nil {
				s.constraintPolicyMutex.Unlock()
				s.log.Error(err, "node-not-found-to-fit-pod")
				continue
			}
			if node == nil {
				s.constraintPolicyMutex.Unlock()
				s.log.V(1).Info("node-assignment-deferred-to-planner", "pod", p.Name)
				continue
			}
			err = s.bindPod(p, node)
			if err != nil {
				s.constraintPolicyMutex.Unlock()
				s.log.Error(err, "failed-to-bind-pod")
				continue
			}
			s.constraintPolicyMutex.Unlock()
			message := fmt.Sprintf("Placed pod [%s/%s] on %s\n", p.Namespace, p.Name, node.Name)
			err = s.emitEvent(p, message)
			if err != nil {
				s.log.Error(err, "failed-to-emit-event")
				continue
			}
		case <-s.quit:
			return nil
		}
	}
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

	pod, err := s.defaultPlanner.GetClientset().CoreV1().Pods(data.Namespace).Get(data.Name, metav1.GetOptions{})
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
		//add back the pod to the podqueue
		s.log.V(1).Info("pod-requeue-add", "pod", pod.Name, "num-requeues", numRequeues)
		if s.options.Extender {
			return true
		}
		s.podQueue <- data
		return false
	}
	s.log.V(1).Info("pod-requeue-requeues-exceeded", "pod", pod.Name)
	return false
}

func (s *ConstraintPolicyScheduler) listenForPodRequeueEvents() {
	defer s.podRequeueQueue.ShutDown()
	go wait.Until(s.requeueWorker, time.Second*5, s.quit)
	<-s.quit
}

func (s *ConstraintPolicyScheduler) requeueWorker() {
	for s.processRequeueEvents() {
	}
}

// requeue pod into the worker queue
func (s *ConstraintPolicyScheduler) requeue(pod *v1.Pod) bool {
	s.podRequeueMutex.Lock()
	if _, ok := s.podRequeueMap[ObjectMeta{Name: pod.Name, Namespace: pod.Namespace}]; !ok {
		s.podRequeueMap[ObjectMeta{Name: pod.Name, Namespace: pod.Namespace}] = struct{}{}
	}
	s.log.V(1).Info("pod-requeue", "pod", pod.Name, "namespace", pod.Namespace)
	s.podRequeueQueue.AddRateLimited(pod)
	s.podRequeueMutex.Unlock()
	if !s.options.Extender {
		return false
	}
	// retry from a rate limited queue for scheduler extender
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
	if !s.options.Extender {
		if pod.Spec.SchedulerName != SchedulerName {
			return false
		}
	}
	return true
}

func (s *ConstraintPolicyScheduler) findFit(pod *v1.Pod, eligibleNodes []v1.Node) (*v1.Node, error) {
retry:
	nodeInstance, err := s.defaultPlanner.FindBestNode(pod, eligibleNodes)
	if err == ErrNoOffersAvailable {
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
			goto retry
		}
		return nil, err
	}
	if err != nil {
		s.log.Error(err, "nodes-not-found", "pod", pod.Name)
		s.log.V(1).Info("requeuing-for-retry", "pod", pod.Name)
		if !s.requeue(pod) {
			return nil, err
		}
		goto retry
	}
	// if the pod was requeued, then cleanup the entry from requeue map
	s.requeueFixup(pod)
	s.log.V(1).Info("found-matching", "node", nodeInstance.Name, "pod", pod.Name)
	return nodeInstance, nil
}

// scheduler extender function to find the best node for the pod
func (s *ConstraintPolicyScheduler) FindBestNodes(pod *v1.Pod, feasibleNodes []v1.Node) ([]v1.Node, error) {
	if !s.options.Extender {
		return nil, fmt.Errorf("Extender config not set. Find best nodes failed for pod %s.", pod.Name)
	}
	s.log.V(1).Info("find-best-nodes", "pod", pod.Name)
	s.constraintPolicyMutex.Lock()
	defer s.constraintPolicyMutex.Unlock()
	if node, err := s.findFit(pod, feasibleNodes); err != nil {
		// if no offers are matched for the pod, return the existing feasible nodes
		if err == ErrNoOffersAvailable {
			return feasibleNodes, nil
		}
		s.log.V(1).Info("scheduler-extender", "no-nodes-available-to-schedule-pod", pod.Name)
		return []v1.Node{}, nil
	} else {
		if node == nil {
			s.log.V(1).Info("scheduler-extender", "pod-waiting-for-planner-assignment", pod.Name)
			return []v1.Node{}, nil
		}
		message := fmt.Sprintf("Placing pod [%s/%s] on %s\n", pod.Namespace, pod.Name, node.Name)
		err = s.emitEvent(pod, message)
		if err != nil {
			s.log.Error(err, "failed-to-emit-event")
		}
		return []v1.Node{*node}, nil
	}
}

func (s *ConstraintPolicyScheduler) bindPod(p *v1.Pod, node *v1.Node) error {
	return s.defaultPlanner.GetClientset().CoreV1().Pods(p.Namespace).Bind(&v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
		},
		Target: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       node.Name,
		},
	})
}

func (s *ConstraintPolicyScheduler) emitEvent(p *v1.Pod, message string) error {
	timestamp := time.Now().UTC()
	_, err := s.defaultPlanner.GetClientset().CoreV1().Events(p.Namespace).Create(&v1.Event{
		Count:          1,
		Message:        message,
		Reason:         "Scheduled",
		LastTimestamp:  metav1.NewTime(timestamp),
		FirstTimestamp: metav1.NewTime(timestamp),
		Type:           "Normal",
		Source: v1.EventSource{
			Component: SchedulerName,
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
	})
	if err != nil {
		return err
	}
	return nil
}

func init() {
	rand.Seed(time.Now().Unix())
}
