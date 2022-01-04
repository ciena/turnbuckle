package scheduler

import (
	"context"
	"errors"
	"fmt"
	constraintv1alpha1 "github.com/ciena/turnbuckle/apis/constraint/v1alpha1"
	constraint_policy_client "github.com/ciena/turnbuckle/internal/pkg/constraint-policy-client"
	"github.com/ciena/turnbuckle/pkg/types"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrNoOffersAvailable = fmt.Errorf("no offers available to schedule pods")
	ErrNoNodesFound      = fmt.Errorf("no eligible nodes found")
)

type ConstraintPolicySchedulerPlanner struct {
	options                ConstraintPolicySchedulerPlannerOptions
	clientset              kubernetes.Interface
	constraintPolicyClient constraint_policy_client.ConstraintPolicyClient
	log                    logr.Logger
	nodeLister             listersv1.NodeLister
	quit                   chan struct{}
	nodeQueue              chan *v1.Node
	podUpdateQueue         workqueue.RateLimitingInterface
	podToNodeMap           map[ObjectMeta]string
	constraintPolicyMutex  sync.Mutex
}

type ConstraintPolicySchedulerPlannerOptions struct{}

type ObjectMeta struct {
	Name      string
	Namespace string
}

type NodeAndCost struct {
	Node string
	Cost int64
}

type constraintPolicyOffer struct {
	offer         *constraintv1alpha1.ConstraintPolicyOffer
	peerToNodeMap map[ObjectMeta]string
	peerNodeNames []string
}

type workWrapper struct {
	work func() error
}

func NewPlannerWithPodCallbacks(options ConstraintPolicySchedulerPlannerOptions, clientset kubernetes.Interface,
	constraintPolicyClient constraint_policy_client.ConstraintPolicyClient,
	log logr.Logger,
	addPodCallback func(pod *v1.Pod),
	updatePodCallback func(old *v1.Pod, new *v1.Pod),
	deletePodCallback func(pod *v1.Pod)) *ConstraintPolicySchedulerPlanner {
	constraintPolicySchedulerPlanner := &ConstraintPolicySchedulerPlanner{
		options:                options,
		clientset:              clientset,
		constraintPolicyClient: constraintPolicyClient,
		log:                    log,
		nodeQueue:              make(chan *v1.Node, 100),
		podUpdateQueue:         workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		quit:                   make(chan struct{}),
		podToNodeMap:           make(map[ObjectMeta]string),
	}

	addFunc := func(obj interface{}) {
		pod, ok := obj.(*v1.Pod)
		if !ok {
			log.V(1).Info("this-is-not-a-pod")
			return
		}
		if addPodCallback != nil {
			addPodCallback(pod)
		}
	}

	updateFunc := func(oldObj interface{}, newObj interface{}) {
		oldPod, ok := oldObj.(*v1.Pod)
		if !ok {
			return
		}
		newPod, ok := newObj.(*v1.Pod)
		if !ok {
			return
		}
		constraintPolicySchedulerPlanner.handlePodUpdate(oldPod, newPod)
		if updatePodCallback != nil {
			updatePodCallback(oldPod, newPod)
		}
	}

	deleteFunc := func(obj interface{}) {
		pod, ok := obj.(*v1.Pod)
		if !ok {
			return
		}
		constraintPolicySchedulerPlanner.handlePodDelete(pod)
		if deletePodCallback != nil {
			deletePodCallback(pod)
		}
	}

	constraintPolicySchedulerPlanner.nodeLister = initInformers(
		clientset,
		log,
		constraintPolicySchedulerPlanner.quit,
		constraintPolicySchedulerPlanner.nodeQueue,
		addFunc,
		updateFunc,
		deleteFunc,
	)

	go constraintPolicySchedulerPlanner.listenForNodeEvents()
	go constraintPolicySchedulerPlanner.listenForPodUpdateEvents()

	return constraintPolicySchedulerPlanner
}

func NewPlanner(options ConstraintPolicySchedulerPlannerOptions, clientset *kubernetes.Clientset,
	constraintPolicyClient constraint_policy_client.ConstraintPolicyClient, log logr.Logger) *ConstraintPolicySchedulerPlanner {
	return NewPlannerWithPodCallbacks(options, clientset, constraintPolicyClient, log, nil, nil, nil)
}

func getEligibleNodes(nodeLister listersv1.NodeLister) ([]*v1.Node, error) {
	nodes, err := nodeLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	var eligibleNodes, preferNoScheduleNodes []*v1.Node
	for _, node := range nodes {
		preferNoSchedule, noSchedule := false, false
		for _, taint := range node.Spec.Taints {
			if taint.Effect == v1.TaintEffectPreferNoSchedule {
				preferNoSchedule = true
			} else if taint.Effect == v1.TaintEffectNoSchedule || taint.Effect == v1.TaintEffectNoExecute {
				noSchedule = true
			}
		}
		if !noSchedule {
			eligibleNodes = append(eligibleNodes, node)
		} else if preferNoSchedule {
			preferNoScheduleNodes = append(preferNoScheduleNodes, node)
		}
	}
	if len(eligibleNodes) == 0 {
		return preferNoScheduleNodes, nil
	} else {
		return eligibleNodes, nil
	}
}

func getEligibleNodeNames(nodeLister listersv1.NodeLister) ([]string, error) {
	eligibleNodeList, err := getEligibleNodes(nodeLister)
	if err != nil {
		return nil, err
	}
	eligibleNodes := make([]string, len(eligibleNodeList))
	for i, eligibleNode := range eligibleNodeList {
		eligibleNodes[i] = eligibleNode.Name
	}
	return eligibleNodes, nil
}

func initInformers(clientset kubernetes.Interface, log logr.Logger, quit chan struct{}, nodeQueue chan *v1.Node,
	addFunc func(obj interface{}), updateFunc func(oldObj interface{}, newObj interface{}), deleteFunc func(obj interface{})) listersv1.NodeLister {
	factory := informers.NewSharedInformerFactory(clientset, 0)
	nodeInformer := factory.Core().V1().Nodes()
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node, ok := obj.(*v1.Node)
			if !ok {
				log.V(1).Info("this-is-not-a-node")
				return
			}
			log.V(1).Info("new-node-added", "node", node.GetName())
			nodeQueue <- node
		},
	})

	podInformer := factory.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	})

	factory.Start(quit)
	return nodeInformer.Lister()
}

func (s *ConstraintPolicySchedulerPlanner) GetClientset() kubernetes.Interface {
	return s.clientset
}

func (s *ConstraintPolicySchedulerPlanner) getEligibleNodesAndNodeNames() ([]*v1.Node, []string, error) {
	nodeRefs, err := getEligibleNodes(s.nodeLister)
	if err != nil {
		return nil, nil, err
	}
	nodeNames := make([]string, len(nodeRefs))
	for i, n := range nodeRefs {
		nodeNames[i] = n.Name
	}
	return nodeRefs, nodeNames, nil
}

func (s *ConstraintPolicySchedulerPlanner) handlePodUpdate(oldPod *v1.Pod, newPod *v1.Pod) {
	s.constraintPolicyMutex.Lock()
	defer s.constraintPolicyMutex.Unlock()
	if oldPod.Status.Phase != v1.PodRunning && newPod.Status.Phase == v1.PodRunning {
		delete(s.podToNodeMap, ObjectMeta{Name: newPod.Name, Namespace: newPod.Namespace})
	} else if oldPod.GetDeletionTimestamp() == nil && newPod.GetDeletionTimestamp() != nil {
		if err := s.releaseUnderlayPath(newPod); err != nil {
			s.log.V(1).Info("handle-pod-update-enqueue-event-on-failure", "pod", newPod.Name)
			s.podUpdateQueue.AddRateLimited(newPod)
		}
	}
}

func (s *ConstraintPolicySchedulerPlanner) releaseUnderlayPath(pod *v1.Pod) error {
	finalizers := pod.GetFinalizers()
	if finalizers == nil {
		return nil
	}
	// check for finalizers if any to release underlay path
	finalizerPrefix := "constraint.ciena.com/remove-underlay_"
	underlayFinalizers := []string{}
	newFinalizers := []string{}
	for _, finalizer := range finalizers {
		if strings.HasPrefix(finalizer, finalizerPrefix) {
			underlayFinalizers = append(underlayFinalizers, finalizer)
		} else {
			newFinalizers = append(newFinalizers, finalizer)
		}
	}
	if len(underlayFinalizers) == 0 {
		return nil
	}
	// update the pod finalizer with the new set
	pod.ObjectMeta.Finalizers = newFinalizers
	if _, err := s.clientset.CoreV1().Pods(pod.Namespace).Update(context.Background(), pod, metav1.UpdateOptions{}); err != nil {
		s.log.Error(err, "underlay-path-release-pod-update-failed", "pod", pod.Name)
		return err
	} else {
		s.log.V(1).Info("underlay-path-release-pod-update-success", "pod", pod.Name)
	}
	underlayController, err := s.lookupUnderlayController()
	if err != nil {
		s.log.Error(err, "underlay-lookup-failed")
		return err
	}
	for _, finalizer := range underlayFinalizers {
		pathId := finalizer[len(finalizerPrefix):]
		s.log.V(1).Info("underlay-path-release", "pod", pod.Name, "path", pathId)
		if err := underlayController.Release(pathId); err != nil {
			s.log.Error(err, "underlay-path-release-failed", "path", pathId)
		} else {
			s.log.V(1).Info("underlay-path-release-success", "path", pathId)
		}
	}

	return nil
}

func (s *ConstraintPolicySchedulerPlanner) handlePodDelete(pod *v1.Pod) {
	s.constraintPolicyMutex.Lock()
	defer s.constraintPolicyMutex.Unlock()
	delete(s.podToNodeMap, ObjectMeta{Name: pod.Name, Namespace: pod.Namespace})
}

func ParseDuration(durationStrings ...string) ([]time.Duration, error) {
	durations := make([]time.Duration, len(durationStrings))
	for i, ds := range durationStrings {
		duration, err := time.ParseDuration(ds)
		if err != nil {
			return nil, err
		}
		durations[i] = duration
	}
	return durations, nil
}

func (s *ConstraintPolicySchedulerPlanner) FindNodeLister(node string) (*v1.Node, error) {
	nodes, err := getEligibleNodes(s.nodeLister)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if n.Name == node {
			return n, nil
		}
	}
	return nil, fmt.Errorf("Cound not find node lister instance for node %s", node)
}

func (s *ConstraintPolicySchedulerPlanner) Stop() {
	close(s.quit)
}

func matchesLabelSelector(labelSelector *metav1.LabelSelector, pod *v1.Pod) (bool, error) {
	set, err := metav1.LabelSelectorAsMap(labelSelector)
	if err != nil {
		return false, err
	}
	return labels.Set(set).AsSelector().Matches(labels.Set(pod.Labels)), nil
}

func (s *ConstraintPolicySchedulerPlanner) getPeerEndpoints(endpointLabels labels.Set) (map[ObjectMeta]string, error) {
	endpoints, err := s.clientset.CoreV1().Endpoints("").List(context.Background(),
		metav1.ListOptions{LabelSelector: endpointLabels.String()})
	if err != nil {
		return nil, err
	}
	// this is a 1:1 as endpoing name is stored for network telemetry info
	endpointNodeMap := make(map[ObjectMeta]string)
	for _, ep := range endpoints.Items {
		endpointNodeMap[ObjectMeta{Name: ep.Name, Namespace: ep.Namespace}] = ep.Name
	}
	return endpointNodeMap, nil
}

func (s *ConstraintPolicySchedulerPlanner) getPeerPods(podLabels labels.Set) (map[ObjectMeta]string, error) {
	pods, err := s.clientset.CoreV1().Pods("").List(context.Background(),
		metav1.ListOptions{LabelSelector: podLabels.String()})
	if err != nil {
		return nil, err
	}
	podNodeMap := make(map[ObjectMeta]string)
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodFailed || pod.DeletionTimestamp != nil {
			continue
		}
		var podNodeName string
		if nodeName, err := s.GetNodeName(pod); err == nil {
			podNodeName = nodeName
		} else {
			podNodeName = ""
		}
		podNodeMap[ObjectMeta{Name: pod.Name, Namespace: pod.Namespace}] = podNodeName
	}

	return podNodeMap, nil
}

func (s *ConstraintPolicySchedulerPlanner) getPeers(selector *constraintv1alpha1.ConstraintPolicyOfferTarget) (map[ObjectMeta]string, error) {
	if selector.Kind == "Pod" && selector.LabelSelector != nil {
		if set, err := metav1.LabelSelectorAsMap(selector.LabelSelector); err != nil {
			s.log.Error(err, "error-getting-label-selector")
			return nil, err
		} else {
			return s.getPeerPods(labels.Set(set))
		}
	}
	if selector.Kind == "Endpoint" && selector.LabelSelector != nil {
		if set, err := metav1.LabelSelectorAsMap(selector.LabelSelector); err != nil {
			s.log.Error(err, "error-getting-label-selector-for-endpoints")
			return nil, err
		} else {
			return s.getPeerEndpoints(labels.Set(set))
		}
	}
	return nil, nil
}

func getPeerNodeNames(peerToNodeMap map[ObjectMeta]string) []string {
	nodeNames := make([]string, 0, len(peerToNodeMap))
	visitedMap := make(map[string]struct{})
	for _, node := range peerToNodeMap {
		if node == "" {
			continue
		}
		if _, ok := visitedMap[node]; !ok {
			nodeNames = append(nodeNames, node)
			visitedMap[node] = struct{}{}
		}
	}
	return nodeNames
}

func mergePeers(p1, p2 map[ObjectMeta]string) map[ObjectMeta]string {
	p3 := make(map[ObjectMeta]string)
	for k, v := range p1 {
		p3[k] = v
	}
	for k, v := range p2 {
		p3[k] = v
	}
	return p3
}

func (s *ConstraintPolicySchedulerPlanner) getPolicyOffers(pod *v1.Pod) ([]constraintPolicyOffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	offers, err := s.constraintPolicyClient.ListConstraintPolicyOffers(ctx, pod.Namespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var offerList []constraintPolicyOffer
	for _, offer := range offers.Items {
		var peerToNodeMap map[ObjectMeta]string
		var peerNodeNames []string
		var matched bool
		peers := []*constraintv1alpha1.ConstraintPolicyOfferTarget{}
		for _, target := range offer.Spec.Targets {
			var targetMatch bool
			if target.Kind == "Pod" && target.LabelSelector != nil {
				if match, err := matchesLabelSelector(target.LabelSelector, pod); err != nil {
					s.log.Error(err, "error-matching-source-label-selector", "offer", offer.Name)
					continue
				} else {
					targetMatch = match
				}
			}
			if !targetMatch {
				peers = append(peers, target)
			} else {
				matched = targetMatch
			}
		}
		if matched {
			for _, peer := range peers {
				if peer.LabelSelector != nil {
					if peerMap, err := s.getPeers(peer); err != nil {
						s.log.Error(err, "error-getting-peers", "offer", offer.Name)
					} else {
						peerToNodeMap = mergePeers(peerMap, peerToNodeMap)
					}
				}
			}
			peerNodeNames = getPeerNodeNames(peerToNodeMap)
			offerList = append(offerList, constraintPolicyOffer{
				offer:         offer,
				peerToNodeMap: peerToNodeMap,
				peerNodeNames: peerNodeNames,
			})
		}
	}
	if len(offerList) == 0 {
		return nil, ErrNoOffersAvailable
	}
	return offerList, nil
}

func (s *ConstraintPolicySchedulerPlanner) lookupUnderlayController() (UnderlayController, error) {
	svcs, err := s.clientset.CoreV1().Services("").List(context.Background(),
		metav1.ListOptions{LabelSelector: "constraint.ciena.com/underlay-controller"})
	if err != nil {
		return nil, err
	}
	if len(svcs.Items) == 0 {
		return nil, errors.New("underlay-controller service not found")
	}
	return &underlayController{
		Log:     s.log.WithName("underlay-controller"),
		Service: svcs.Items[0],
	}, nil
}

func (s *ConstraintPolicySchedulerPlanner) lookupRuleProvider(name, namespace string) (RuleProvider, error) {
	svcs, err := s.clientset.CoreV1().Services("").List(context.Background(),
		metav1.ListOptions{LabelSelector: fmt.Sprintf("constraint.ciena.com/provider-%s", name)})
	if err != nil {
		return nil, err
	}
	if len(svcs.Items) == 0 {
		return nil, errors.New("not-found")
	}
	return &ruleProvider{
		Log:         s.log.WithName("rule-provider").WithName(name),
		ProviderFor: name,
		Service:     svcs.Items[0],
	}, nil
}

func mergeOfferCost(m1 map[string]int64, m2 map[string]int64) map[string]int64 {
	m3 := make(map[string]int64)
	for k, v := range m1 {
		if v2, ok := m2[k]; ok {
			m3[k] = (v + v2) / 2
		}
	}
	return m3
}

func mergeNodeCost(m1 map[string][]int64, m2 map[string][]int64) map[string][]int64 {
	m3 := make(map[string][]int64)
	for k, v1 := range m1 {
		if v2, ok := m2[k]; ok {
			// concat m1[k] and m2[k] values on intersect
			m3[k] = append(m3[k], v1...)
			m3[k] = append(m3[k], v2...)
		}
	}

	return m3
}

func mergeNodeAllocationCost(m1 map[NodeAndCost]string, m2 map[NodeAndCost]string) map[NodeAndCost]string {
	m3 := make(map[NodeAndCost]string)
	for nc := range m1 {
		if m2Id, ok := m2[nc]; ok {
			// we can use both as id is in both the lists
			m3[nc] = m2Id
		}
	}

	return m3
}

func mergeRules(existingRules []*constraintv1alpha1.ConstraintPolicyRule, newRules []*constraintv1alpha1.ConstraintPolicyRule) []*constraintv1alpha1.ConstraintPolicyRule {
	presentMap := make(map[string]struct{})
	for _, r := range existingRules {
		presentMap[r.Name] = struct{}{}
	}
	for _, r := range newRules {
		if _, ok := presentMap[r.Name]; !ok {
			existingRules = append(existingRules, r)
		}
	}
	return existingRules
}

func getAggregate(values []int64) int64 {
	var sum int64
	for i := range values {
		sum += values[i]
	}
	if len(values) > 1 {
		sum = sum / int64(len(values))
	}
	return sum
}

func filterOutInfiniteCost(nodeAndCost []NodeAndCost) []NodeAndCost {
	out := []NodeAndCost{}
	for _, nc := range nodeAndCost {
		if nc.Cost >= 0 {
			out = append(out, nc)
		}
	}
	return out
}

func (s *ConstraintPolicySchedulerPlanner) getEndpointCost(src types.Reference, policyRules []*constraintv1alpha1.ConstraintPolicyRule,
	eligibleNodes []string, peerNodeNames []string) (map[string]int64, []*constraintv1alpha1.ConstraintPolicyRule, error) {
	ruleToCostMap := make(map[string][]NodeAndCost)
	matchedRules := []*constraintv1alpha1.ConstraintPolicyRule{}
	for _, rule := range policyRules {
		if provider, err := s.lookupRuleProvider(rule.Name, ""); err != nil {
			s.log.Error(err, "error-looking-up-rule-provider", "rule", rule.Name)
		} else {
			if nodeAndCost, err := provider.EndpointCost(src, eligibleNodes, peerNodeNames,
				rule.Request, rule.Limit); err != nil {
				s.log.Error(err, "error-getting-endpoint-cost", "rule", rule.Name)
			} else {
				ruleToCostMap[rule.Name] = filterOutInfiniteCost(nodeAndCost)
				if len(nodeAndCost) > 0 {
					matchedRules = append(matchedRules, rule)
				}
			}
		}
	}
	var mergedNodeCostMap map[string][]int64
	for _, nodeAndCost := range ruleToCostMap {
		nodeCostMap := make(map[string][]int64)
		for _, nc := range nodeAndCost {
			nodeCostMap[nc.Node] = append(nodeCostMap[nc.Node], nc.Cost)
		}
		if mergedNodeCostMap != nil {
			mergedNodeCostMap = mergeNodeCost(mergedNodeCostMap, nodeCostMap)
		} else {
			mergedNodeCostMap = nodeCostMap
		}
	}
	if mergedNodeCostMap == nil {
		return nil, nil, errors.New("unable to get endpoint cost")
	}
	// for each node compute the aggregated cost now from the intersects
	mergedNodeAggregateCostMap := make(map[string]int64)
	for node, values := range mergedNodeCostMap {
		mergedNodeAggregateCostMap[node] = getAggregate(values)
	}
	return mergedNodeAggregateCostMap, matchedRules, nil
}

func (s *ConstraintPolicySchedulerPlanner) getUnderlayCost(src types.Reference, eligibleNodes []string, peerNodeMap map[ObjectMeta]string,
	peerNodeNames []string, rules []*constraintv1alpha1.ConstraintPolicyRule) (map[string]int64, map[NodeAndCost]string, error) {
	// this will make a grpc to underlay and get the cost
	underlayController, err := s.lookupUnderlayController()
	if err != nil {
		s.log.Error(err, "underlay-controller-lookup-failed")
		return nil, nil, err
	}

	nodeOffers, err := underlayController.Discover(eligibleNodes, peerNodeNames, rules)
	if err != nil {
		s.log.Error(err, "underlay-controller-path-discovery-failed")
		return nil, nil, err
	}

	nodeCostMap := make(map[string]int64)
	nodeAllocationPathMap := make(map[NodeAndCost]string)

	for _, offer := range nodeOffers {
		nodeCostMap[offer.node] = offer.cost
		nodeAllocationPathMap[NodeAndCost{Node: offer.node, Cost: offer.cost}] = offer.id
	}

	return nodeCostMap, nodeAllocationPathMap, nil
}

func (s *ConstraintPolicySchedulerPlanner) getUnderlayCostForOffers(matchingOffers []constraintPolicyOffer, pod *v1.Pod, eligibleNodes []string,
	offerToRulesMap map[string][]*constraintv1alpha1.ConstraintPolicyRule) (map[string]int64, map[NodeAndCost]string, error) {

	var offerCostMap map[string]int64
	var nodeAllocationPathMap map[NodeAndCost]string

	for _, matchingOffer := range matchingOffers {
		rules, ok := offerToRulesMap[matchingOffer.offer.Name]
		if !ok {
			continue
		}
		if len(rules) == 0 {
			s.log.V(1).Info("get-underlay-cost-no-rules-found", "offer", matchingOffer.offer.Name)
			continue
		}
		var peerNodeNames []string
		if len(matchingOffer.peerNodeNames) == 0 {
			peerNodeNames = eligibleNodes
		} else {
			peerNodeNames = matchingOffer.peerNodeNames
		}
		nodeCostMap, allocationPathMap, err := s.getUnderlayCost(types.Reference{Name: pod.Name, Kind: "Pod"}, eligibleNodes,
			matchingOffer.peerToNodeMap, peerNodeNames, rules)
		if err != nil {
			s.log.Error(err, "error-getting-underlay-cost", "offer", matchingOffer.offer.Name)
			continue
		}
		if offerCostMap != nil {
			offerCostMap = mergeOfferCost(offerCostMap, nodeCostMap)
		} else {
			offerCostMap = nodeCostMap
		}
		if nodeAllocationPathMap != nil {
			nodeAllocationPathMap = mergeNodeAllocationCost(nodeAllocationPathMap, allocationPathMap)
		} else {
			nodeAllocationPathMap = allocationPathMap
		}
	}

	if offerCostMap == nil || len(offerCostMap) == 0 {
		return nil, nil, fmt.Errorf("No underlay cost obtained for offers")
	}

	return offerCostMap, nodeAllocationPathMap, nil
}

func (s *ConstraintPolicySchedulerPlanner) getNodeWithBestCost(nodeCostMap map[string]int64, applyFilter func(string) bool) (NodeAndCost, error) {

	var nodeCostList []NodeAndCost
	for n, c := range nodeCostMap {
		nodeCostList = append(nodeCostList, NodeAndCost{Node: n, Cost: c})
	}

	if len(nodeCostList) == 0 {
		return NodeAndCost{}, ErrNoNodesFound
	}

	//sort the list based on cost
	sort.Slice(nodeCostList, func(i, j int) bool {
		return nodeCostList[i].Cost < nodeCostList[j].Cost
	})

	// filter out nodes that don't satisfy the filter
	for _, nodeAndCost := range nodeCostList {

		if applyFilter != nil && !applyFilter(nodeAndCost.Node) {
			continue
		}

		return nodeAndCost, nil
	}

	return NodeAndCost{}, ErrNoNodesFound
}

func (s *ConstraintPolicySchedulerPlanner) getPodCandidateNodes(pod *v1.Pod, eligibleNodes []string, offers []constraintPolicyOffer) (map[string]int64, map[NodeAndCost]string, error) {
	var offerCostMap map[string]int64
	offerToRulesMap := make(map[string][]*constraintv1alpha1.ConstraintPolicyRule)
	nodeAllocationPathMap := make(map[NodeAndCost]string)
	for _, matchingOffer := range offers {
		var policyRules []*constraintv1alpha1.ConstraintPolicyRule
		for _, policyName := range matchingOffer.offer.Spec.Policies {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			policy, err := s.constraintPolicyClient.GetConstraintPolicy(ctx, matchingOffer.offer.Namespace, string(policyName), metav1.GetOptions{})
			cancel()
			if err != nil {
				s.log.Error(err, "error-getting-policy", "policy", string(policyName))
				continue
			}
			policyRules = mergeRules(policyRules, policy.Spec.Rules)
		}
		nodeCostMap, matchedRules, err := s.getEndpointCost(types.Reference{Name: pod.Name, Kind: "Pod"},
			policyRules, eligibleNodes, matchingOffer.peerNodeNames)
		if err != nil {
			s.log.Error(err, "error-getting-endpoint-cost", "offer", matchingOffer.offer.Name)
			continue
		}
		offerToRulesMap[matchingOffer.offer.Name] = matchedRules

		if offerCostMap != nil {
			offerCostMap = mergeOfferCost(offerCostMap, nodeCostMap)
		} else {
			offerCostMap = nodeCostMap
		}
	}

	if offerCostMap == nil {
		return nil, nil, fmt.Errorf("No offers were matched for pod %s", pod.Name)
	}

	if len(offerCostMap) == 0 {
		// if no nodes were found from ruleprovider, we go to the underlay
		if costMap, allocationPathMap, err := s.getUnderlayCostForOffers(offers, pod, eligibleNodes, offerToRulesMap); err != nil {
			s.log.Error(err, "get-underlay-cost-for-offers-failed")
			return nil, nil, err
		} else {
			offerCostMap = costMap
			nodeAllocationPathMap = allocationPathMap
		}
	}

	return offerCostMap, nodeAllocationPathMap, nil
}

func (s *ConstraintPolicySchedulerPlanner) getNodeCost(pod *v1.Pod, eligibleNodes []string, offers []constraintPolicyOffer) (map[string]int64, error) {
	var offerCostMap map[string]int64

	for _, matchingOffer := range offers {
		var policyRules []*constraintv1alpha1.ConstraintPolicyRule
		for _, policyName := range matchingOffer.offer.Spec.Policies {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			policy, err := s.constraintPolicyClient.GetConstraintPolicy(ctx, matchingOffer.offer.Namespace, string(policyName), metav1.GetOptions{})
			cancel()
			if err != nil {
				s.log.Error(err, "error-getting-policy", "policy", string(policyName))
				continue
			}
			policyRules = mergeRules(policyRules, policy.Spec.Rules)
		}

		nodeCostMap, _, err := s.getEndpointCost(types.Reference{Name: pod.Name, Kind: "Pod"},
			policyRules, eligibleNodes, matchingOffer.peerNodeNames)
		if err != nil {
			s.log.Error(err, "error-getting-endpoint-cost", "offer", matchingOffer.offer.Name)
			continue
		}

		if offerCostMap != nil {
			offerCostMap = mergeOfferCost(offerCostMap, nodeCostMap)
		} else {
			offerCostMap = nodeCostMap
		}
	}

	if offerCostMap == nil {
		return nil, fmt.Errorf("Could not get endpoint cost for pod %s", pod.Name)
	}

	return offerCostMap, nil
}

func (s *ConstraintPolicySchedulerPlanner) Start() {
	go s.listenForNodeEvents()
	go s.listenForPodUpdateEvents()
}

func (s *ConstraintPolicySchedulerPlanner) listenForNodeEvents() {
	for {
		select {
		case <-s.nodeQueue:
			//nothing to do for now

		case <-s.quit:
			return
		}
	}
}

// called with constraintpolicymutex held.
func (s *ConstraintPolicySchedulerPlanner) getPodNode(pod v1.Pod) (string, error) {
	if node, ok := s.podToNodeMap[ObjectMeta{Name: pod.Name, Namespace: pod.Namespace}]; !ok {
		return "", fmt.Errorf("Cannot find pod %s node", pod.Name)
	} else {
		return node, nil
	}
}

// called with constraintpolicymutex held.
func (s *ConstraintPolicySchedulerPlanner) setPodNode(pod v1.Pod, nodeName string) {
	s.podToNodeMap[ObjectMeta{Name: pod.Name, Namespace: pod.Namespace}] = nodeName
}

func (s *ConstraintPolicySchedulerPlanner) FindFitRandom(pod *v1.Pod, nodes []*v1.Node) (*v1.Node, error) {
	if len(nodes) == 0 {
		if eligibleNodes, _, err := s.getEligibleNodesAndNodeNames(); err != nil {
			return nil, err
		} else {
			nodes = eligibleNodes
		}
	}
	return nodes[rand.Intn(len(nodes))], nil
}

// Fast version. look up internal cache.
// look up for the node in the lister cache if pod host ip is not set.
func (s *ConstraintPolicySchedulerPlanner) GetNodeName(pod v1.Pod) (string, error) {
	if pod.Status.HostIP == "" {
		return s.getPodNode(pod)
	}
	nodes, err := getEligibleNodes(s.nodeLister)
	if err != nil {
		return "", err
	}
	for _, node := range nodes {
		for _, nodeAddr := range node.Status.Addresses {
			if nodeAddr.Address == pod.Status.HostIP {
				return node.Name, nil
			}
		}
	}
	return "", fmt.Errorf("Pod ip %s not found in node lister cache", pod.Status.HostIP)
}

func (s *ConstraintPolicySchedulerPlanner) ToNodeName(endpoint types.Reference) (string, error) {
	if endpoint.Kind != "Pod" {
		return endpoint.Name, nil
	}
	pods, err := s.clientset.CoreV1().Pods(endpoint.Namespace).List(context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return "", err
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodFailed || pod.DeletionTimestamp != nil {
			continue
		}
		if pod.Name == endpoint.Name {
			if nodeName, err := s.GetNodeName(pod); err == nil {
				return nodeName, nil
			} else {
				return "", err
			}
		}
	}

	return "", fmt.Errorf("Unable to find nodename for pod %s", endpoint.Name)
}

func (s *ConstraintPolicySchedulerPlanner) processUpdateEvents() bool {
	// get blocks till there is an item
	item, quit := s.podUpdateQueue.Get()
	if quit {
		return false
	}
	defer s.podUpdateQueue.Done(item)
	s.processUpdate(item)
	return true
}

func (s *ConstraintPolicySchedulerPlanner) AddRateLimited(work func() error) {
	s.podUpdateQueue.AddRateLimited(&workWrapper{work: work})
}

func (s *ConstraintPolicySchedulerPlanner) processUpdate(item interface{}) {
	forgetItem := true
	defer func() {
		if forgetItem {
			s.podUpdateQueue.Forget(item)
		}
	}()
	var data *v1.Pod
	switch item.(type) {
	case *v1.Pod:
		data = item.(*v1.Pod)
	case *workWrapper:
		workWrapped := item.(*workWrapper)
		if err := workWrapped.work(); err != nil {
			forgetItem = false
			s.log.V(1).Info("pod-update-work-failed", "numrequeues", s.podUpdateQueue.NumRequeues(item))
			s.podUpdateQueue.AddRateLimited(item)
			return
		}
		return
	default:
		s.log.V(1).Info("pod-update-item-not-a-pod-or-a-function")
		return
	}

	// get the latest version of the item
	pod, err := s.clientset.CoreV1().Pods(data.Namespace).Get(context.Background(), data.Name, metav1.GetOptions{})
	if err != nil {
		return
	}
	if pod.GetDeletionTimestamp() == nil {
		s.log.V(1).Info("pod-update-no-deletion-timestamp", "pod", pod.Name)
		return
	}

	if pod.GetFinalizers() == nil {
		s.log.V(1).Info("pod-update-no-finalizers-found", "pod", pod.Name)
	}

	// try to release the underlay path.
	// requeue the update back on failure
	err = s.releaseUnderlayPath(pod)
	if err != nil {
		forgetItem = false
		s.log.V(1).Info("pod-update-release-underlay-path-failed", "pod", pod.Name, "numrequeues", s.podUpdateQueue.NumRequeues(item))
		s.podUpdateQueue.AddRateLimited(item)
		return
	}
	s.log.V(1).Info("pod-update-release-underlay-path-success", "pod", pod.Name)
	return
}

func (s *ConstraintPolicySchedulerPlanner) listenForPodUpdateEvents() {
	defer s.podUpdateQueue.ShutDown()
	go wait.Until(s.updateWorker, time.Second*5, s.quit)
	<-s.quit
}

func (s *ConstraintPolicySchedulerPlanner) updateWorker() {
	for s.processUpdateEvents() {
	}
}

func (s *ConstraintPolicySchedulerPlanner) setPodFinalizer(pod *v1.Pod, pathId string) error {
	podFinalizer := "constraint.ciena.com/remove-underlay_" + pathId
	for _, finalizer := range pod.ObjectMeta.Finalizers {
		if finalizer == podFinalizer {
			// already exists
			return nil
		}
	}
	pod.ObjectMeta.Finalizers = append(pod.ObjectMeta.Finalizers, podFinalizer)
	if _, err := s.clientset.CoreV1().Pods(pod.Namespace).Update(context.Background(), pod, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (s *ConstraintPolicySchedulerPlanner) podFitsNode(pod *v1.Pod, nodename string) bool {
	return true
}

func (s *ConstraintPolicySchedulerPlanner) findFit(pod *v1.Pod, eligibleNodes []*v1.Node, eligibleNodeNames []string) (*v1.Node, error) {

	if len(eligibleNodes) == 0 {
		s.log.V(1).Info("no-eligible-nodes-found", "pod", pod.Name)
		return nil, fmt.Errorf("no-eligible-nodes-found-for-pod-%s", pod.Name)
	}

	offers, err := s.getPolicyOffers(pod)
	if err != nil {
		return nil, err
	}

	nodeCostMap, nodeAllocationPathMap, err := s.getPodCandidateNodes(pod, eligibleNodeNames, offers)
	if err != nil {
		s.log.Error(err, "nodes-not-found", "pod", pod.Name)
		return nil, err
	}

	s.log.V(1).Info("pod-node-assignment", "pod", pod.Name, "node-candidates", nodeCostMap)

	// find if the node was assigned to the neighbor pod
	matchedNode, err := s.getNodeWithBestCost(nodeCostMap, func(nodeName string) bool {
		return s.podFitsNode(pod, nodeName)
	})

	if err != nil {
		s.log.Error(err, "eligible-nodes-not-found", "pod", pod.Name)
		return nil, err
	}

	s.log.V(1).Info("pod-node-assignment", "pod", pod.Name, "node", matchedNode.Node, "cost", matchedNode.Cost)

	nodeInstance, err := s.FindNodeLister(matchedNode.Node)
	if err != nil {
		s.log.V(1).Info("node-instance-not-found-in-lister-cache", "node", matchedNode.Node)
		s.log.V(1).Info("random-assignment", "pod", pod.Name)
		return s.FindFitRandom(pod, eligibleNodes)
	}

	s.setPodNode(*pod, nodeInstance.Name)

	// we check if we have to allocate a path to the underlay for this node
	if underlayPathId, ok := nodeAllocationPathMap[matchedNode]; ok {

		if underlayController, err := s.lookupUnderlayController(); err != nil {
			s.log.Error(err, "underlay-controller-lookup-failed", "for-node-and-costr", matchedNode)
		} else {
			if err := underlayController.Allocate(underlayPathId); err != nil {
				s.log.V(1).Info("underlay-controller-allocate-failed", "path-id", underlayPathId)
			} else {
				s.log.V(1).Info("underlay-controller-allocate-success", "path-id", underlayPathId)
			}
			if err := s.setPodFinalizer(pod, underlayPathId); err != nil {
				s.log.Error(err, "set-pod-finalizer-failed", "pod", pod.Name, "path", underlayPathId)
			} else {
				s.log.V(1).Info("set-pod-finalizer-success", "pod", pod.Name, "path", underlayPathId)
			}
		}
	}

	return nodeInstance, nil
}

// scheduler function to find the best node for the pod.
func (s *ConstraintPolicySchedulerPlanner) FindBestNode(pod *v1.Pod, feasibleNodes []*v1.Node) (*v1.Node, error) {
	var nodeNames []string
	if len(feasibleNodes) == 0 {
		if eligibleNodes, nodes, err := s.getEligibleNodesAndNodeNames(); err != nil {
			return nil, err
		} else {
			feasibleNodes = eligibleNodes
			nodeNames = nodes
		}
	} else {
		nodeNames = make([]string, len(feasibleNodes))
		for i, n := range feasibleNodes {
			nodeNames[i] = n.Name
		}
	}
	s.log.V(1).Info("find-best-node", "pod", pod.Name)

	s.constraintPolicyMutex.Lock()
	defer s.constraintPolicyMutex.Unlock()

	if node, err := s.findFit(pod, feasibleNodes, nodeNames); err != nil {
		// if no offers are matched for the pod, return the existing feasible nodes
		return nil, err
	} else {
		return node, nil
	}
}
