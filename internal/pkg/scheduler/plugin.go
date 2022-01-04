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
	Name = "ConstraintPolicyScheduler"
)

type ConstraintPolicyScheduling struct {
	scheduler *ConstraintPolicyScheduler
	fh        framework.Handle
}

var _ framework.PostFilterPlugin = &ConstraintPolicyScheduling{}

func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	var log logr.Logger

	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	log = zapr.NewLogger(zapLog)

	clientset := handle.ClientSet()
	kubeconfig := handle.KubeConfig()

	constraintPolicyClient, err := constraint_policy_client.New(kubeconfig, log.WithName("constraint-policy-client"))
	if err != nil {
		log.Error(err, "Error initializing constraint policy client interface")
		return nil, err
	}

	constraintPolicyScheduler := NewScheduler(ConstraintPolicySchedulerOptions{
		NumRetriesOnFailure: schedulerConfig.NumRetriesOnFailure,
		MinDelayOnFailure:   schedulerConfig.MinDelayOnFailure,
		MaxDelayOnFailure:   schedulerConfig.MaxDelayOnFailure,
		FallbackOnNoOffers:  schedulerConfig.FallbackOnNoOffers,
	},
		clientset, handle, constraintPolicyClient,
		log.WithName("constraint-policy").WithName("scheduler"))

	constraintPolicyScheduling := &ConstraintPolicyScheduling{fh: handle, scheduler: constraintPolicyScheduler}
	return constraintPolicyScheduling, nil
}

func (c *ConstraintPolicyScheduling) Name() string {
	return Name
}

func (c *ConstraintPolicyScheduling) PostFilter(ctx context.Context,
	state *framework.CycleState, pod *v1.Pod,
	filteredNodeStatusMap framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	var eligibleNodes []*v1.Node

	allNodes, err := c.fh.SnapshotSharedLister().NodeInfos().List()
	if err != nil {
		return nil, framework.AsStatus(err)
	}

	for _, nodeInfo := range allNodes {
		if filteredNodeStatusMap[nodeInfo.Node().Name].Code() == framework.Success {
			eligibleNodes = append(eligibleNodes, nodeInfo.Node())
		}
	}

	if len(eligibleNodes) == 0 {
		c.scheduler.log.V(1).Info("post-filter-no-nodes-eligible")
		return nil, framework.NewStatus(framework.Unschedulable)
	}

	node, err := c.scheduler.FindBestNode(pod, eligibleNodes)

	if err != nil {
		return nil, framework.AsStatus(err)
	}

	if node == nil {
		return nil, framework.NewStatus(framework.Unschedulable)
	}

	return &framework.PostFilterResult{NominatedNodeName: node.Name}, framework.NewStatus(framework.Success)
}
