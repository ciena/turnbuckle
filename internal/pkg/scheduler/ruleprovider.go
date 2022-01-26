package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/ciena/turnbuckle/pkg/apis/ruleprovider"
	"github.com/ciena/turnbuckle/pkg/types"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
)

// RuleProvider declares the methods that can be called against a rule provider.
type RuleProvider interface {
	// EndpointCost calculates the cost of use each of the eligible nodes.
	EndpointCost(src *types.Reference, eligibleNodes []string,
		peerNodes []string, request, limit string) ([]NodeAndCost, error)
}

type ruleProvider struct {
	Log         logr.Logger
	ProviderFor string
	Service     corev1.Service
	DialOptions []grpc.DialOption
	CallTimeout time.Duration
}

// EndpointCost calculates the cost of use each of the eligible nodes.
func (p *ruleProvider) EndpointCost(
	src *types.Reference,
	eligibleNodes []string,
	peerNodes []string,
	request, limit string) ([]NodeAndCost, error) {
	p.Log.V(1).Info("endpointcost", "namespace", p.Service.Namespace, "name", p.Service.Name)
	dns := fmt.Sprintf("%s.%s.svc.cluster.local:5309", p.Service.Name, p.Service.Namespace)

	dctx, dcancel := context.WithTimeout(context.Background(), p.CallTimeout)
	defer dcancel()

	conn, err := grpc.DialContext(dctx, dns, p.DialOptions...)
	if err != nil {
		return nil, fmt.Errorf("error connecting to rule provider: %w", err)
	}

	// nolint:errcheck
	defer conn.Close()

	client := ruleprovider.NewRuleProviderClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), p.CallTimeout)
	defer cancel()

	source := ruleprovider.Target{
		Cluster:    src.Cluster,
		Kind:       src.Kind,
		ApiVersion: src.APIVersion,
		Name:       src.Name,
		Namespace:  src.Namespace,
	}

	req := ruleprovider.EndpointCostRequest{
		Source:        &source,
		EligibleNodes: eligibleNodes,
		PeerNodes:     peerNodes,
		Rule: &ruleprovider.PolicyRule{
			Name:    p.ProviderFor,
			Request: request,
			Limit:   limit,
		},
	}

	resp, err := client.EndpointCost(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("error calculation endpoint cost: %w", err)
	}

	if len(resp.NodeAndCost) == 0 {
		return nil, ErrNoNodesFound
	}

	nodeAndCost := make([]NodeAndCost, len(resp.NodeAndCost))
	for i, nc := range resp.NodeAndCost {
		nodeAndCost[i] = NodeAndCost{
			Node: nc.Node,
			Cost: nc.Cost,
		}
	}

	return nodeAndCost, nil
}
