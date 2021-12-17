package podpredicates

import (
	"reflect"
	"sort"
	"strings"

	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	podutil "sigs.k8s.io/descheduler/pkg/descheduler/pod"
)

// PodIsADuplicate checks if the given pod is a duplicate of another pod on the same node
// A pod is said to be a duplicate of other if both of them are from same creator, kind and are within the same
// namespace, and have at least one container with the same image.
func PodIsADuplicate(
	client clientset.Interface,
	pod *v1.Pod,
	node *v1.Node) (bool, error) {

	pods, err := podutil.ListPodsOnANode(client, node)
	if err != nil {
		return false, err
	}

	givenPodContainerKeys := getPodContainerKeys(pod)
	for _, p := range pods {
		// skip terminating pods
		if p.GetDeletionTimestamp() != nil {
			continue
		}
		podContainerKeys := getPodContainerKeys(p)
		if reflect.DeepEqual(givenPodContainerKeys, podContainerKeys) {
			// given pod is a duplicate of another pod on this node
			return true, nil
		}
	}

	return false, nil
}

func getPodContainerKeys(pod *v1.Pod) []string {
	ownerRefList := podutil.OwnerRef(pod)
	podContainerKeys := make([]string, 0, len(ownerRefList)*len(pod.Spec.Containers))
	for _, ownerRef := range ownerRefList {
		for _, container := range pod.Spec.Containers {
			// Namespace/Kind/Name should be unique for the cluster.
			// We also consider the image, as 2 pods could have the same owner but serve different purposes
			// So any non-unique Namespace/Kind/Name/Image pattern is a duplicate pod.
			s := strings.Join([]string{pod.ObjectMeta.Namespace, ownerRef.Kind, ownerRef.Name, container.Image}, "/")
			podContainerKeys = append(podContainerKeys, s)
		}
	}
	sort.Strings(podContainerKeys)
	return podContainerKeys
}
