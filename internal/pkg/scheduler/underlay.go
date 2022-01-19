package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	constraintv1alpha1 "github.com/ciena/turnbuckle/pkg/apis/constraint/v1alpha1"
	"github.com/ciena/turnbuckle/pkg/apis/underlay"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
)

type nodeOffer struct {
	node    string
	cluster string
	id      string
	cost    int64
	expires time.Time
}

type UnderlayController interface {
	Discover(eligibleNodes []string, peerNodes []string, rules []*constraintv1alpha1.ConstraintPolicyRule) ([]nodeOffer, error)
	Allocate(pathId string) error
	Release(pathId string) error
}

type underlayController struct {
	Log     logr.Logger
	Service corev1.Service
}

func (c *underlayController) Discover(eligibleNodes []string, peerNodes []string, rules []*constraintv1alpha1.ConstraintPolicyRule) ([]nodeOffer, error) {
	var gopts []grpc.DialOption
	c.Log.V(1).Info("discover", "namespace", c.Service.Namespace, "name", c.Service.Name)
	dns := fmt.Sprintf("%s.%s.svc.cluster.local:9999", c.Service.Name, c.Service.Namespace)

	gopts = append(gopts, grpc.WithInsecure())
	conn, err := grpc.Dial(dns, gopts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := underlay.NewUnderlayControllerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var policyRules []*underlay.PolicyRule
	// unused, so just send empty
	for range rules {
		policyRules = append(policyRules, &underlay.PolicyRule{})
	}
	req := underlay.DiscoverRequest{
		Rules:         policyRules,
		EligibleNodes: eligibleNodes,
		PeerNodes:     peerNodes,
	}
	resp, err := client.Discover(ctx, &req)
	if err != nil {
		return nil, err
	}
	if len(resp.Offers) == 0 {
		return nil, errors.New("No nodes found that can be setup by the underlay controller")
	}

	var nodeOffers []nodeOffer
	for _, offer := range resp.Offers {
		nodeOffers = append(nodeOffers, nodeOffer{
			node:    offer.Node.Name,
			cluster: offer.Node.Cluster,
			cost:    offer.Cost,
			id:      offer.Id,
		})
	}
	return nodeOffers, nil
}

func (c *underlayController) Allocate(pathId string) error {
	var gopts []grpc.DialOption
	c.Log.V(1).Info("allocatepaths", "namespace", c.Service.Namespace, "name", c.Service.Name)
	dns := fmt.Sprintf("%s.%s.svc.cluster.local:9999", c.Service.Name, c.Service.Namespace)

	gopts = append(gopts, grpc.WithInsecure())
	conn, err := grpc.Dial(dns, gopts...)
	if err != nil {
		return fmt.Errorf("unable to connect to underlay controller: %w", err)
	}
	defer conn.Close()

	client := underlay.NewUnderlayControllerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := underlay.AllocateRequest{Id: pathId}
	_, err = client.Allocate(ctx, &req)
	if err != nil {
		return fmt.Errorf("unable to allocate underlay: %w", err)
	}
	return nil
}

func (c *underlayController) Release(pathId string) error {
	var gopts []grpc.DialOption
	c.Log.V(1).Info("releasepaths", "namespace", c.Service.Namespace, "name", c.Service.Name)
	dns := fmt.Sprintf("%s.%s.svc.cluster.local:9999", c.Service.Name, c.Service.Namespace)

	gopts = append(gopts, grpc.WithInsecure())
	conn, err := grpc.Dial(dns, gopts...)
	if err != nil {
		return fmt.Errorf("unable to connect to underlay controller: %w", err)
	}
	defer conn.Close()

	client := underlay.NewUnderlayControllerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := underlay.ReleaseRequest{Id: pathId}
	_, err = client.Release(ctx, &req)
	if err != nil {
		return fmt.Errorf("unable to release client connection: %w", err)
	}
	return nil
}
