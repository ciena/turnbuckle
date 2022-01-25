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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// PluginArgs defines the parameters for ConstraintPolicyScheduling plugin
type ConstraintPolicySchedulingArgs struct {
	metav1.TypeMeta      `json:",inline"`
	Debug                bool   `json:"debug,omitEmpty"`
	MinDelayOnFailure    string `json:"minDelayOnFailure,omitEmpty"`
	MaxDelayOnFailure    string `json:"maxDelayOnFailure,omitEmpty"`
	NumRetriesOnFailure  int    `json:"numRetriesOnFailure,omitEmpty"`
	FallbackOnNoOffers   bool   `json:"fallbackOnNoOffers,omitEmpty"`
	RetryOnNoOffers      bool   `json:"retryOnNoOffers,omitEmpty"`
	RequeuePeriod        string `json:"requeuePeriod,omitEmpty"`
	PlannerNodeQueueSize uint   `json:"plannerNodeQueueSize,omitEmpty"`
}

func DefaultConstraintPolicySchedulerConfig() *ConstraintPolicySchedulerOptions {
	// nolint:gomnd
	return &ConstraintPolicySchedulerOptions{
		Debug:                true,
		NumRetriesOnFailure:  3,
		MinDelayOnFailure:    30 * time.Second,
		MaxDelayOnFailure:    time.Minute,
		FallbackOnNoOffers:   false,
		RetryOnNoOffers:      false,
		RequeuePeriod:        5 * time.Second,
		PlannerNodeQueueSize: 300,
	}
}

func parsePluginConfig(pluginConfig *ConstraintPolicySchedulingArgs, defaultConfig *ConstraintPolicySchedulerOptions) *ConstraintPolicySchedulerOptions {
	var config ConstraintPolicySchedulerOptions = *defaultConfig

	config.Debug = pluginConfig.Debug
	config.FallbackOnNoOffers = pluginConfig.FallbackOnNoOffers
	config.RetryOnNoOffers = pluginConfig.RetryOnNoOffers

	if pluginConfig.NumRetriesOnFailure > 0 {
		config.NumRetriesOnFailure = pluginConfig.NumRetriesOnFailure
	}

	minDelay, err := time.ParseDuration(pluginConfig.MinDelayOnFailure)
	if err == nil && minDelay > time.Duration(0) {
		config.MinDelayOnFailure = minDelay
	}

	maxDelay, err := time.ParseDuration(pluginConfig.MaxDelayOnFailure)
	if err == nil && maxDelay > time.Duration(0) {
		config.MaxDelayOnFailure = maxDelay
	}

	if config.MinDelayOnFailure > config.MaxDelayOnFailure {
		config.MinDelayOnFailure = config.MaxDelayOnFailure / 2
	}

	requeuePeriod, err := time.ParseDuration(pluginConfig.RequeuePeriod)
	if err == nil && requeuePeriod > time.Duration(0) {
		config.RequeuePeriod = requeuePeriod
	}

	if pluginConfig.PlannerNodeQueueSize > 0 {
		config.PlannerNodeQueueSize = pluginConfig.PlannerNodeQueueSize
	}

	return &config
}
