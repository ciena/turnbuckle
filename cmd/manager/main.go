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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	controllers "github.com/ciena/turnbuckle/controllers/constraint"
	cpv1 "github.com/ciena/turnbuckle/pkg/apis/constraint/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

//nolint:gochecknoglobals
var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	leaderElectionID          = "64e1845a-5786-11ec-bf63-0242ac130002.ciena.com"
	defaultRPCTimeout         = 10 * time.Second
	defaultEvaluationInterval = 10 * time.Second
	defaultErrorInterval      = 10 * time.Second
)

//nolint:gochecknoinits,wsl
func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(cpv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

type configSpec struct {
	ConfigurationFile              string
	EnableLeaderElection           bool
	OfferEvaluationInterval        time.Duration
	OfferEvaluationErrorInterval   time.Duration
	BindingEvaluationInterval      time.Duration
	BindingEvaluationErrorInterval time.Duration
	RPCTimeout                     time.Duration
	ShowVersion                    bool
	ShowVersionAsJSON              bool
}

func main() {
	var config configSpec

	flag.StringVar(&config.ConfigurationFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	flag.BoolVar(&config.EnableLeaderElection,
		"enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active "+
			"controller manager.")
	flag.DurationVar(&config.OfferEvaluationInterval,
		"offer-evaluation-interval", defaultEvaluationInterval,
		"The interval at which offer sources and destinations are "+
			"evalauted to create bindings.")
	flag.DurationVar(&config.OfferEvaluationErrorInterval,
		"offer-evaluation-error-interval", defaultErrorInterval,
		"The interval until an offer is re-evaluated when an error is "+
			"seen during its processing")
	flag.DurationVar(&config.BindingEvaluationInterval,
		"binding-evaluation-interval", defaultEvaluationInterval,
		"The interval at which bindings are evaluated.")
	flag.DurationVar(&config.BindingEvaluationErrorInterval,
		"binding-evaluation-error-interval", defaultErrorInterval,
		"The interval until at which bindings are re-evaluated when an error is "+
			"seen during its processing")
	flag.DurationVar(&config.RPCTimeout,
		"rpc-timeout", defaultRPCTimeout,
		"Timeout for RPC calls to external services")
	flag.BoolVar(&config.ShowVersion,
		"version", false,
		"Display version information and exit")
	flag.BoolVar(&config.ShowVersionAsJSON,
		"json", false,
		"If displaying version, do so a JSON object.")

	//nolint:exhaustruct
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	if config.ShowVersion {
		if config.ShowVersionAsJSON {
			//nolint:errchkjson
			bytes, _ := json.Marshal(controllers.Version())
			fmt.Fprintln(os.Stdout, string(bytes))
		} else {
			fmt.Fprintln(os.Stdout, controllers.Version().String())
		}

		os.Exit(0)
	}

	var err error

	//nolint:exhaustruct
	options := ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   config.EnableLeaderElection,
		LeaderElectionID: leaderElectionID,
	}

	if config.ConfigurationFile != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(config.ConfigurationFile))
		if err != nil {
			setupLog.Error(err, "unable to load the config file")
			os.Exit(1)
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ConstraintPolicyReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ConstraintPolicy")
		os.Exit(1)
	}

	if err = (&controllers.ConstraintPolicyOfferReconciler{
		Client:                  mgr.GetClient(),
		Scheme:                  mgr.GetScheme(),
		EvaluationInterval:      config.OfferEvaluationInterval,
		EvaluationErrorInterval: config.OfferEvaluationErrorInterval,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ConstraintPolicyOffer")
		os.Exit(1)
	}

	if err = (&controllers.ConstraintPolicyBindingReconciler{
		Client:                  mgr.GetClient(),
		Scheme:                  mgr.GetScheme(),
		EvaluationInterval:      config.BindingEvaluationInterval,
		EvaluationErrorInterval: config.BindingEvaluationErrorInterval,
		RPCTimeout:              config.RPCTimeout,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ConstraintPolicyBinding")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
