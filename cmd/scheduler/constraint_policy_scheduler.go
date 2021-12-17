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
	"flag"
	"fmt"
	constraint_policy_client "github.com/ciena/turnbuckle/internal/pkg/constraint-policy-client"
	"github.com/ciena/turnbuckle/internal/pkg/scheduler"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
)

type configSpec struct {
	MetricsAddr           string
	EnableLeaderElection  bool
	Extender              bool
	Debug                 bool
	MinDelayOnFailure     time.Duration
	MaxDelayOnFailure     time.Duration
	NumRetriesOnFailure   int
	FallbackOnNoOffers    bool
	RetryOnNoOffers       bool
	CheckForDuplicatePods bool
}

var (
	constraintPolicyScheduler *scheduler.ConstraintPolicyScheduler
	scheme                    = runtime.NewScheme()
	setupLog                  = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	// +kubebuilder:scaffold:scheme
}

func k8sClient(kubeconfig string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if kubeconfig == "" {
		if config, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
	} else {
		if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func main() {
	var config configSpec
	flag.StringVar(&config.MetricsAddr,
		"metrics-addr", ":8080",
		"The address the metric endpoint binds to.")
	flag.BoolVar(&config.EnableLeaderElection,
		"enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&config.Extender,
		"extender", false,
		"Run as an extender to the default k8s scheduler")
	flag.BoolVar(&config.Debug,
		"debug", true,
		"Display debug logging messages")
	flag.IntVar(&config.NumRetriesOnFailure, "num-retries-on-failure", 0,
		"Number of retries to schedule the pod on scheduling failures. Use <= 0 to retry indefinitely.")
	flag.DurationVar(&config.MinDelayOnFailure, "min-delay-on-failure", time.Minute,
		"The minimum delay interval for rescheduling pods on failures.")
	flag.DurationVar(&config.MaxDelayOnFailure, "max-delay-on-failure", time.Minute*3,
		"The maximum delay interval before rescheduling pods on failures.")
	flag.BoolVar(&config.FallbackOnNoOffers,
		"schedule-on-no-offers", false,
		"Schedule a pod if no offers are found")
	flag.BoolVar(&config.RetryOnNoOffers,
		"retry-on-no-offers", true,
		"Keep retrying to schedule a pod if no offers are found")
	flag.BoolVar(&config.CheckForDuplicatePods,
		"check-for-duplicate-pods", false,
		"Skip scheduling pods on nodes where there are duplicates of the pod.")

	flag.Parse()

	var kubeConfig string
	if fl := flag.Lookup("kubeconfig"); fl != nil {
		kubeConfig = fl.Value.String()
	}
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
	ctrl.SetLogger(log)
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: config.MetricsAddr,
		Port:               9443,
		LeaderElection:     config.EnableLeaderElection,
		LeaderElectionID:   "c9efbb85.ciena.io",
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	mgrStart := func() {
		log.Info("starting-manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			log.Error(err, "manager-start-failure")
			os.Exit(1)
		}
	}
	clientset, err := k8sClient(kubeConfig)
	if err != nil {
		log.Error(err, "Error initializing k8s client interface")
		os.Exit(1)
	}
	constraintPolicyClient, err := constraint_policy_client.New(kubeConfig, log.WithName("constraint-policy-client"))
	if err != nil {
		log.Error(err, "Error initializing constraint policy client interface")
		os.Exit(1)
	}
	if config.Extender {
		if config.NumRetriesOnFailure <= 0 {
			config.NumRetriesOnFailure = 2
		}
		if config.MinDelayOnFailure > time.Duration(time.Second*30) {
			config.MinDelayOnFailure = time.Duration(time.Second * 30)
		}
		if config.MaxDelayOnFailure > time.Duration(time.Minute) {
			config.MaxDelayOnFailure = time.Duration(time.Minute)
		}
	}
	constraintPolicyScheduler = scheduler.NewScheduler(scheduler.ConstraintPolicySchedulerOptions{
		Extender:              config.Extender,
		NumRetriesOnFailure:   config.NumRetriesOnFailure,
		MinDelayOnFailure:     config.MinDelayOnFailure,
		MaxDelayOnFailure:     config.MaxDelayOnFailure,
		FallbackOnNoOffers:    config.FallbackOnNoOffers,
		CheckForDuplicatePods: config.CheckForDuplicatePods,
	},
		clientset, constraintPolicyClient, log.WithName("constraint-policy").WithName("scheduler"))

	// first start the manager before starting the scheduler
	go mgrStart()

	if !config.Extender {
		constraintPolicyScheduler.Run()
	} else {
		// run in extender mode
		log.Info("scheduler-mode-extender")
		constraintPolicyScheduler.Start()
		router := httprouter.New()
		router.GET("/", Index)
		router.POST("/filter", Filter)
		log.Error(http.ListenAndServe(":8888", router), "listen-and-serve-failed")
		constraintPolicyScheduler.Stop()
	}
}
