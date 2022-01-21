/*
Copyright 2021 Ciena Corporation.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package scheduler

import (
	"context"
	"errors"
	"fmt"

	constraint_policy_client "github.com/ciena/turnbuckle/internal/pkg/constraint-policy-client"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	Name              = "ConstraintPolicyScheduling"
	preFilterStateKey = "PreFilter" + Name
)

type ConstraintPolicyScheduling struct {
	scheduler *ConstraintPolicyScheduler
	fh        framework.Handle
	log       logr.Logger
}

type preFilterState struct {
	node string
}

var (
	_ framework.PreFilterPlugin  = &ConstraintPolicyScheduling{}
	_ framework.ScorePlugin      = &ConstraintPolicyScheduling{}
	_ framework.PostFilterPlugin = &ConstraintPolicyScheduling{}
)

func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	var log logr.Logger

	if config.Debug {
		zapLog, err := zap.NewDevelopment()
		if err != nil {
			panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
		}
		log = zapr.NewLogger(zapLog)
	} else {
		zapLog, err := zap.NewProduction()
		if err != nil {
			panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
		}
		log = zapr.NewLogger(zapLog)
	}

	clientset := handle.ClientSet()
	kubeconfig := handle.KubeConfig()

	constraintPolicyClient, err := constraint_policy_client.New(kubeconfig, log.WithName("constraint-policy-client"))
	if err != nil {
		log.Error(err, "Error initializing constraint policy client interface")
		return nil, err
	}

	constraintPolicyScheduler := NewScheduler(ConstraintPolicySchedulerOptions{
		NumRetriesOnFailure: config.NumRetriesOnFailure,
		MinDelayOnFailure:   config.MinDelayOnFailure,
		MaxDelayOnFailure:   config.MaxDelayOnFailure,
		FallbackOnNoOffers:  config.FallbackOnNoOffers,
		RetryOnNoOffers:     config.RetryOnNoOffers,
	},
		clientset, handle, constraintPolicyClient,
		log.WithName("constraint-policy").WithName("scheduler"))

	pluginLogger := log.WithName("scheduling-plugin")

	constraintPolicyScheduling := &ConstraintPolicyScheduling{
		fh:        handle,
		scheduler: constraintPolicyScheduler,
		log:       pluginLogger,
	}

	pluginLogger.V(1).Info("constraint-policy-scheduling-plugin-initialized")

	return constraintPolicyScheduling, nil
}

func getPreFilterState(cycleState *framework.CycleState) (*preFilterState, error) {
	state, err := cycleState.Read(preFilterStateKey)
	if err != nil {
		return nil, err
	}

	if state == nil {
		return nil, errors.New("assignment-state-nil")
	}

	assignmentState, ok := state.(*preFilterState)
	if !ok {
		return nil, fmt.Errorf("%+v convert to node assignment state error", state)
	}

	return assignmentState, nil
}

func (s *preFilterState) Clone() framework.StateData {
	return s
}

func (c *ConstraintPolicyScheduling) createPreFilterState(ctx context.Context, pod *v1.Pod) (*preFilterState, *framework.Status) {
	allNodes, err := c.fh.SnapshotSharedLister().NodeInfos().List()
	if err != nil {
		return nil, framework.AsStatus(err)
	}

	eligibleNodes := make([]*v1.Node, len(allNodes))

	for i, nodeInfo := range allNodes {
		eligibleNodes[i] = nodeInfo.Node()
	}

	if len(eligibleNodes) == 0 {
		c.log.V(1).Info("pre-filter-no-nodes-eligible")
		return nil, framework.NewStatus(framework.Unschedulable)
	}

	node, err := c.scheduler.FindBestNode(pod, eligibleNodes)
	if err != nil {
		return nil, framework.AsStatus(err)
	}

	if node == nil {
		return nil, framework.NewStatus(framework.Unschedulable)
	}

	return &preFilterState{node: node.Name}, framework.NewStatus(framework.Success)
}

func (c *ConstraintPolicyScheduling) Name() string {
	return Name
}

func (c *ConstraintPolicyScheduling) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	c.log.V(1).Info("prefilter", "pod", pod.Name)

	assignmentState, status := c.createPreFilterState(ctx, pod)
	if !status.IsSuccess() {
		return status
	}

	state.Write(preFilterStateKey, assignmentState)

	return status
}

// PreFilterExtensions returns prefilter extensions, pod add and remove.
func (c *ConstraintPolicyScheduling) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (c *ConstraintPolicyScheduling) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	c.log.V(1).Info("score", "pod", pod.Name, "node", nodeName)

	assignmentState, err := getPreFilterState(state)
	if err != nil {
		return framework.MinNodeScore, framework.NewStatus(framework.Success)
	}

	if assignmentState.node != nodeName {
		return 1, framework.NewStatus(framework.Success)
	}

	c.log.V(1).Info("set-score", "score", framework.MaxNodeScore, "pod", pod.Name, "node", nodeName)

	return framework.MaxNodeScore, framework.NewStatus(framework.Success)
}

func (c *ConstraintPolicyScheduling) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// PostFilter is called when no node can be assigned to the pod.
func (c *ConstraintPolicyScheduling) PostFilter(ctx context.Context,
	state *framework.CycleState, pod *v1.Pod,
	filteredNodeStatusMap framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	c.log.V(1).Info("post-filter", "pod", pod.Name)

	assignmentState, err := getPreFilterState(state)
	if err != nil {
		return nil, framework.AsStatus(err)
	}

	c.log.V(1).Info("post-filter", "nominated-node", assignmentState.node, "pod", pod.Name)

	return &framework.PostFilterResult{NominatedNodeName: assignmentState.node}, framework.NewStatus(framework.Success)
}
