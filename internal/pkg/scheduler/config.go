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

type configSpec struct {
	Debug               bool
	MinDelayOnFailure   time.Duration
	MaxDelayOnFailure   time.Duration
	NumRetriesOnFailure int
	FallbackOnNoOffers  bool
	RetryOnNoOffers     bool
}

var config configSpec = configSpec{
	Debug:               true,
	NumRetriesOnFailure: 3,
	MinDelayOnFailure:   30 * time.Second,
	MaxDelayOnFailure:   time.Minute,
	FallbackOnNoOffers:  false,
	RetryOnNoOffers:     false,
}

func AddFlags(cmd *cobra.Command) *flag.FlagSet {
	var nfs cliflag.NamedFlagSets

	fs := nfs.FlagSet("Constraint policy flags")

	fs.BoolVar(&config.Debug, "debug", config.Debug, "Enable debug logs")
	fs.BoolVar(&config.FallbackOnNoOffers,
		"schedule-on-no-offers", config.FallbackOnNoOffers,
		"Schedule a pod if no offers are found")
	fs.BoolVar(&config.RetryOnNoOffers,
		"retry-on-no-offers", config.RetryOnNoOffers,
		"Keep retrying to schedule a pod if no offers are found")
	fs.DurationVar(&config.MinDelayOnFailure, "min-delay-on-failure",
		config.MinDelayOnFailure, "The minimum delay interval for rescheduling pods on failures.")
	fs.DurationVar(&config.MaxDelayOnFailure, "max-delay-on-failure",
		config.MaxDelayOnFailure, "The maximum delay interval before rescheduling pods on failures.")
	fs.IntVar(&config.NumRetriesOnFailure, "num-retries-on-failure", config.NumRetriesOnFailure,
		"Number of retries to schedule the pod on scheduling failures. Use <= 0 to retry indefinitely.")

	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())

	cliflag.SetUsageAndHelpFunc(cmd, nfs, cols)

	return fs
}
