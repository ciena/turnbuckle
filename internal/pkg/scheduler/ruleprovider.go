package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ciena/turnbuckle/apis/ruleprovider"
	graph "github.com/ciena/turnbuckle/internal/pkg/graph"
	"github.com/ciena/turnbuckle/pkg/types"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
)

type RuleProvider interface {
	EndpointCost(src types.Reference, eligibleNodes []string, peerNodes []string, request, limit string) ([]graph.NodeAndCost, error)
}

type ruleProvider struct {
	Log         logr.Logger
	ProviderFor string
	Service     corev1.Service
}

func (p *ruleProvider) EndpointCost(src types.Reference, eligibleNodes []string, peerNodes []string, request, limit string) ([]graph.NodeAndCost, error) {
	var gopts []grpc.DialOption
	p.Log.V(1).Info("endpointcost", "namespace", p.Service.Namespace, "name", p.Service.Name)
	dns := fmt.Sprintf("%s.%s.svc.cluster.local:5309", p.Service.Name, p.Service.Namespace)

	gopts = append(gopts, grpc.WithInsecure())
	conn, err := grpc.Dial(dns, gopts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := ruleprovider.NewRuleProviderClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
		return nil, err
	}
	if len(resp.NodeAndCost) == 0 {
		return nil, errors.New("No node found")
	}

	var nodeAndCost []graph.NodeAndCost
	for _, nc := range resp.NodeAndCost {
		nodeAndCost = append(nodeAndCost, graph.NodeAndCost{
			Node: nc.Node,
			Cost: nc.Cost,
		})
	}
	return nodeAndCost, nil
}
