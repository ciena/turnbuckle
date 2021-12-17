package nsm

import (
	"context"
	"github.com/go-logr/logr"
	networkservice_v1alpha1 "github.com/networkservicemesh/networkservicemesh/k8s/pkg/apis/networkservice/v1alpha1"
	networkservices "github.com/networkservicemesh/networkservicemesh/k8s/pkg/networkservice/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type NetworkServiceClient interface {
	ListNetworkServices(ctx context.Context, namespace string, opts metav1.ListOptions) (*networkservice_v1alpha1.NetworkServiceList, error)
	GetNetworkService(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*networkservice_v1alpha1.NetworkService, error)
	DecomposeNetworkServices(ctx context.Context, namespace, mode string, label labels.Set) (map[string][]NetworkServiceEntry, error)
}

type networkServiceClient struct {
	clientset    *networkservices.Clientset
	k8sClientset *kubernetes.Clientset
	log          logr.Logger
}

type serviceChain struct {
	root  string
	mode  string
	chain map[string]string
}

type NetworkServiceEntry struct {
	Name string
	Pods []v1.Pod
}

const (
	defaultMode = "End"
)

func New(kubeConfig string, log logr.Logger) (NetworkServiceClient, error) {
	var config *rest.Config
	var err error
	if kubeConfig == "" {
		if config, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
	} else {
		if config, err = clientcmd.BuildConfigFromFlags("", kubeConfig); err != nil {
			return nil, err
		}
	}
	clientset, err := networkservices.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	k8sClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &networkServiceClient{clientset: clientset, k8sClientset: k8sClientset, log: log}, nil
}

func (n *networkServiceClient) ListNetworkServices(ctx context.Context, namespace string, opts metav1.ListOptions) (*networkservice_v1alpha1.NetworkServiceList, error) {
	networkServices, err := n.clientset.NetworkservicemeshV1alpha1().NetworkServices(namespace).List(opts)
	if err != nil {
		return nil, err
	}
	return networkServices, nil
}

func (n *networkServiceClient) GetNetworkService(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*networkservice_v1alpha1.NetworkService, error) {
	networkservice, err := n.clientset.NetworkservicemeshV1alpha1().NetworkServices(namespace).Get(name, opts)
	if err != nil {
		return nil, err
	}
	return networkservice, nil
}

func (n *networkServiceClient) DecomposeNetworkServices(ctx context.Context, namespace, mode string, labels labels.Set) (map[string][]NetworkServiceEntry, error) {
	networkServices, err := n.ListNetworkServices(ctx, namespace, metav1.ListOptions{LabelSelector: labels.String()})
	if err != nil {
		return nil, err
	}
	networkServiceMap := make(map[string][]NetworkServiceEntry)
	for _, networkService := range networkServices.Items {
		n.log.V(1).Info("decompose-nsm", "service", networkService.Name)
		networkServiceMap[networkService.Name] = n.decomposeNetworkService(&networkService, mode)
	}
	return networkServiceMap, nil
}

func (n *networkServiceClient) decomposeNetworkService(networkService *networkservice_v1alpha1.NetworkService, mode string) []NetworkServiceEntry {
	serviceChain := newServiceChain(mode)
	for _, m := range networkService.Spec.Matches {
		source, destination := "ANY", ""
		for k, v := range m.SourceSelector {
			if k == "app" {
				source = v
			}
		}
		for _, route := range m.Routes {
			for k, v := range route.DestinationSelector {
				if k == "app" {
					destination = v
				}
			}
		}
		serviceChain.add(source, destination)
	}
	chains := serviceChain.build()
	nse := make([]NetworkServiceEntry, 0, len(chains))
	for _, chain := range chains {
		pods, err := n.k8sClientset.CoreV1().Pods("").List(metav1.ListOptions{LabelSelector: "app=" + chain})
		if err != nil {
			nse = append(nse, NetworkServiceEntry{Name: chain, Pods: []v1.Pod{}})
		} else {
			nse = append(nse, NetworkServiceEntry{Name: chain, Pods: pods.Items})
		}
	}

	return nse
}

func newServiceChain(mode string) *serviceChain {
	if mode == "" || (mode != "Start" && mode != "End") {
		mode = defaultMode
	}
	return &serviceChain{root: "ANY", mode: mode, chain: make(map[string]string)}
}

func (s *serviceChain) add(src, dst string) error {
	chain := src
	if chain == "" {
		chain = s.root
	}
	s.chain[chain] = dst
	return nil
}

func (s *serviceChain) build() []string {
	var chains []string
	dst := s.chain[s.root]
	for s.chain[dst] != "" {
		chains = append(chains, dst)
		dst = s.chain[dst]
	}
	if dst != "" {
		chains = append(chains, dst)
	}
	switch s.mode {
	case "Start":
		if len(chains) > 0 {
			chains = chains[:1]
		}
	case "End":
	default:
	}

	return chains
}
