package client

import (
	"context"
	"time"

	constraintv1alpha1 "github.com/ciena/turnbuckle/apis/constraint/v1alpha1"
	"github.com/go-logr/logr"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	Scheme         = runtime.NewScheme()
	Codecs         = serializer.NewCodecFactory(Scheme)
	ParameterCodec = runtime.NewParameterCodec(Scheme)
)

type ConstraintPolicyClient interface {
	GetConstraintPolicyBinding(ctx context.Context, namespace, name string, opts v1.GetOptions) (*constraintv1alpha1.ConstraintPolicyBinding, error)
	GetConstraintPolicyOffer(ctx context.Context, namespace, name string, opts v1.GetOptions) (*constraintv1alpha1.ConstraintPolicyOffer, error)
	GetConstraintPolicy(ctx context.Context, namespace, name string, opts v1.GetOptions) (*constraintv1alpha1.ConstraintPolicy, error)
	ListConstraintPolicyBindings(ctx context.Context, namespace string, opts v1.ListOptions) (*constraintv1alpha1.ConstraintPolicyBindingList, error)
	ListConstraintPolicyOffers(ctx context.Context, namespace string, opts v1.ListOptions) (*constraintv1alpha1.ConstraintPolicyOfferList, error)
	ListConstraintPolicies(ctx context.Context, namespace string, opts v1.ListOptions) (*constraintv1alpha1.ConstraintPolicyList, error)
}

type constraintPolicyClient struct {
	client rest.Interface
	log    logr.Logger
}

func init() {
	v1.AddToGroupVersion(Scheme, constraintv1alpha1.GroupVersion)
	constraintv1alpha1.AddToScheme(Scheme)
}

func New(kubeConfig string, log logr.Logger) (ConstraintPolicyClient, error) {
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
	client, err := newConstraintClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return &constraintPolicyClient{client: client, log: log}, nil
}

func newConstraintClientForConfig(c *rest.Config) (rest.Interface, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to create REST client: %w", err)
	}
	return client, nil
}

func setConfigDefaults(config *rest.Config) error {
	gv := constraintv1alpha1.GroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = Codecs.WithoutConversion()
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

func (c *constraintPolicyClient) GetConstraintPolicyBinding(ctx context.Context, namespace, name string, opts v1.GetOptions) (result *constraintv1alpha1.ConstraintPolicyBinding, err error) {
	result = &constraintv1alpha1.ConstraintPolicyBinding{}
	err = c.client.Get().
		Namespace(namespace).
		Resource("constraintpolicybindings").
		Name(name).
		VersionedParams(&opts, ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

func (c *constraintPolicyClient) GetConstraintPolicyOffer(ctx context.Context, namespace, name string, opts v1.GetOptions) (result *constraintv1alpha1.ConstraintPolicyOffer, err error) {
	result = &constraintv1alpha1.ConstraintPolicyOffer{}
	err = c.client.Get().
		Namespace(namespace).
		Resource("constraintpolicyoffers").
		Name(name).
		VersionedParams(&opts, ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

func (c *constraintPolicyClient) GetConstraintPolicy(ctx context.Context, namespace, name string, opts v1.GetOptions) (result *constraintv1alpha1.ConstraintPolicy, err error) {
	result = &constraintv1alpha1.ConstraintPolicy{}
	err = c.client.Get().
		Namespace(namespace).
		Resource("constraintpolicies").
		Name(name).
		VersionedParams(&opts, ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

func (c *constraintPolicyClient) ListConstraintPolicyBindings(ctx context.Context, namespace string, opts v1.ListOptions) (result *constraintv1alpha1.ConstraintPolicyBindingList, err error) {
	result = &constraintv1alpha1.ConstraintPolicyBindingList{}
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	err = c.client.Get().
		Namespace(namespace).
		Resource("constraintpolicybindings").
		VersionedParams(&opts, ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

func (c *constraintPolicyClient) ListConstraintPolicyOffers(ctx context.Context, namespace string, opts v1.ListOptions) (result *constraintv1alpha1.ConstraintPolicyOfferList, err error) {
	result = &constraintv1alpha1.ConstraintPolicyOfferList{}
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	err = c.client.Get().
		Namespace(namespace).
		Resource("constraintpolicyoffers").
		VersionedParams(&opts, ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

func (c *constraintPolicyClient) ListConstraintPolicies(ctx context.Context, namespace string, opts v1.ListOptions) (result *constraintv1alpha1.ConstraintPolicyList, err error) {
	result = &constraintv1alpha1.ConstraintPolicyList{}
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	err = c.client.Get().
		Namespace(namespace).
		Resource("constraintpolicies").
		VersionedParams(&opts, ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}
