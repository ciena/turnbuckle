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
	"time"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"
)

type ConstraintPolicySchedulerConfig struct {
	Debug                bool
	MinDelayOnFailure    time.Duration
	MaxDelayOnFailure    time.Duration
	NumRetriesOnFailure  int
	FallbackOnNoOffers   bool
	RetryOnNoOffers      bool
	RequeuePeriod        time.Duration
	PodQueueSize         uint
	PlannerNodeQueueSize uint
}

func DefaultConstraintPolicySchedulerConfig() *ConstraintPolicySchedulerConfig {
	// nolint:gomnd
	return &ConstraintPolicySchedulerConfig{
		Debug:                true,
		NumRetriesOnFailure:  3,
		MinDelayOnFailure:    30 * time.Second,
		MaxDelayOnFailure:    time.Minute,
		FallbackOnNoOffers:   false,
		RetryOnNoOffers:      false,
		RequeuePeriod:        5 * time.Second,
		PodQueueSize:         300,
		PlannerNodeQueueSize: 300,
	}
}

// AddFlags adds scheduler specific command line flags.
func (config ConstraintPolicySchedulerConfig) AddFlags(cmd *cobra.Command) *flag.FlagSet {
	var nfs cliflag.NamedFlagSets

	defaults := DefaultConstraintPolicySchedulerConfig()

	fs := nfs.FlagSet("Constraint policy flags")

	fs.BoolVar(&config.Debug, "debug", defaults.Debug, "Enable debug logs")
	fs.BoolVar(&config.FallbackOnNoOffers, "schedule-on-no-offers", defaults.FallbackOnNoOffers,
		"Schedule a pod if no offers are found")
	fs.BoolVar(&config.RetryOnNoOffers, "retry-on-no-offers", defaults.RetryOnNoOffers,
		"Keep retrying to schedule a pod if no offers are found")
	fs.DurationVar(&config.MinDelayOnFailure, "min-delay-on-failure",
		defaults.MinDelayOnFailure, "The minimum delay interval for rescheduling pods on failures.")
	fs.DurationVar(&config.MaxDelayOnFailure, "max-delay-on-failure",
		defaults.MaxDelayOnFailure, "The maximum delay interval before rescheduling pods on failures.")
	fs.IntVar(&config.NumRetriesOnFailure, "num-retries-on-failure", defaults.NumRetriesOnFailure,
		"Number of retries to schedule the pod on scheduling failures. Use <= 0 to retry indefinitely.")
	fs.DurationVar(&config.RequeuePeriod, "requeue-period",
		defaults.RequeuePeriod, "How often schedule workers should be requeued.")
	fs.UintVar(&config.PodQueueSize, "pod-queue-size",
		defaults.PodQueueSize, "Size of queue for maintaining incoming requests")
	fs.UintVar(&config.PlannerNodeQueueSize, "planner-node-queue-size",
		defaults.PlannerNodeQueueSize, "Size of queue for maintaining incoming requests")

	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())

	cliflag.SetUsageAndHelpFunc(cmd, nfs, cols)

	return fs
}
