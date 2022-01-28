/*
Copyright 2022 Ciena Corporation.

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

package constraint_test

import (
	"context"
	"strings"
	"testing"

	controllers "github.com/ciena/turnbuckle/controllers/constraint"
	cpv1 "github.com/ciena/turnbuckle/pkg/apis/constraint/v1alpha1"
	ctypes "github.com/ciena/turnbuckle/pkg/types"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestRemoveInvalidBinding(t *testing.T) {
	t.Parallel()

	// Create a pod that will not map to the existing
	// policy bindings
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "sample",
		},
	}

	// Add the scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(cpv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	// Create a fake k8s client and add test data to it
	builder := fake.NewClientBuilder()
	builder.WithScheme(scheme)
	builder.WithObjects(&pod)
	builder.WithLists(&policies, &offers, &bindings)
	con := builder.Build()

	// Query existing Pods in default namespace
	var pods corev1.PodList

	// Validate that we have a single pod
	err := con.List(context.TODO(),
		&pods,
		client.InNamespace("default"))

	assert.Nil(t, err, "error listing pods")
	assert.Equal(t, 1, len(pods.Items), "wrong number of pre-existing pods")

	var before cpv1.ConstraintPolicyBindingList

	// Validate that we have a single pre-existing binding
	err = con.List(context.TODO(),
		&before,
		client.InNamespace("default"))

	assert.Nil(t, err, "error listing bindings")
	assert.Equal(t, 1, len(before.Items), "wrong number of pre-existing bindings")

	// create an instance of policy offer reconciler
	cpo := &controllers.ConstraintPolicyOfferReconciler{
		Client: con,
		Scheme: scheme,
	}

	// Invoke the policy offer reconciler, which should remove unused bindings
	req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "gold-offer"}}
	_, err = cpo.Reconcile(context.TODO(), req)

	assert.Nil(t, err, "error invoking policy offer reconciler")

	var left cpv1.ConstraintPolicyBindingList

	// Validate binding was removed
	err = con.List(context.TODO(),
		&left,
		client.InNamespace("default"))

	assert.Nil(t, err, "error listing left bindings")
	assert.Equal(t, 0, len(left.Items), "bindings still exist")
}

func TestBindingCreated(t *testing.T) {
	t.Parallel()

	// Create a pod that will not map to the existing
	// policy bindings
	pods := corev1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ListMeta: metav1.ListMeta{},
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "server",
					Labels: map[string]string{
						"app": "hello-server",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "client",
					Labels: map[string]string{
						"app": "hello-client",
					},
				},
			},
		},
	}

	// Add the scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(cpv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	// Create a fake k8s client and add test data to it
	builder := fake.NewClientBuilder()
	builder.WithScheme(scheme)
	builder.WithLists(&pods, &policies, &offers)
	con := builder.Build()

	// Query existing Pods in default namespace
	var curPods corev1.PodList

	// Validate that we have a single pod
	err := con.List(context.TODO(),
		&curPods,
		client.InNamespace("default"))

	assert.Nil(t, err, "error listing pods")
	assert.Equal(t, 2, len(curPods.Items), "wrong number of pre-existing pods")

	var before, left cpv1.ConstraintPolicyBindingList

	// Validate that we have a single pre-existing binding
	err = con.List(context.TODO(),
		&before,
		client.InNamespace("default"))

	assert.Nil(t, err, "error listing bindings")
	assert.Equal(t, 0, len(before.Items), "wrong number of pre-existing bindings")

	// create an instance of policy offer reconciler
	cpo := &controllers.ConstraintPolicyOfferReconciler{
		Client: con,
		Scheme: scheme,
	}

	// Invoke the policy offer reconciler, which should remove unused bindings
	req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "gold-offer"}}
	_, err = cpo.Reconcile(context.TODO(), req)

	assert.Nil(t, err, "error invoking policy offer reconciler")

	// Validate no binding was removed
	err = con.List(context.TODO(),
		&left,
		client.InNamespace("default"))

	assert.Nil(t, err, "error listing left bindings")
	assert.Equal(t, 1, len(left.Items), "incorrent number of bindings left")
	assert.Equal(t, 1, len(left.Items[0].ObjectMeta.Labels), "wrong number of labels on binding")

	val, ok := left.Items[0].ObjectMeta.Labels["constraint.ciena.io/constraintPolicyOffer"]
	assert.Equal(t, ok, true, "label to map to offer does not exist")
	assert.Equal(t, "gold-offer", val, "binding associated with incorrect offer")

	assert.Equal(t, true, strings.HasPrefix(left.Items[0].ObjectMeta.Name, "gold-offer-"), "incorrect binding name")
	assert.Equal(t, 2, len(left.Items[0].Spec.Targets), "wrong number of targets")

	dest, ok := left.Items[0].Spec.Targets["destination"]

	assert.Equal(t, true, ok, "destination not in binding")
	assert.Equal(t, ctypes.Reference{
		Cluster:    "",
		APIVersion: "v1",
		Kind:       "Pod",
		Namespace:  "default",
		Name:       "server",
	},
		dest, "destination reference not correct")

	source, ok := left.Items[0].Spec.Targets["source"]

	assert.Equal(t, true, ok, "source not in binding")
	assert.Equal(t, ctypes.Reference{
		Cluster:    "",
		APIVersion: "v1",
		Kind:       "Pod",
		Namespace:  "default",
		Name:       "client",
	},
		source, "source reference not correct")
}

func TestOfferNotFound(t *testing.T) {
	t.Parallel()

	// Create a pod that will not map to the existing
	// policy bindings
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "sample",
		},
	}

	// Add the scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(cpv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	// Create a fake k8s client and add test data to it
	builder := fake.NewClientBuilder()
	builder.WithScheme(scheme)
	builder.WithObjects(&pod)
	builder.WithLists(&policies, &offers, &bindings)
	con := builder.Build()

	// Query existing Pods in default namespace
	var pods corev1.PodList

	// Validate that we have a single pod
	err := con.List(context.TODO(),
		&pods,
		client.InNamespace("default"))

	assert.Nil(t, err, "error listing pods")
	assert.Equal(t, 1, len(pods.Items), "wrong number of pre-existing pods")

	var before, left cpv1.ConstraintPolicyBindingList

	// Validate that we have a single pre-existing binding
	err = con.List(context.TODO(),
		&before,
		client.InNamespace("default"))

	assert.Nil(t, err, "error listing bindings")
	assert.Equal(t, 1, len(before.Items), "wrong number of pre-existing bindings")

	// create an instance of policy offer reconciler
	cpo := &controllers.ConstraintPolicyOfferReconciler{
		Client: con,
		Scheme: scheme,
	}

	// Invoke the policy offer reconciler, which should remove unused bindings
	req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "does-not-exist"}}
	_, err = cpo.Reconcile(context.TODO(), req)

	assert.Nil(t, err, "error invoking policy offer reconciler")

	// Validate no binding was removed
	err = con.List(context.TODO(),
		&left,
		client.InNamespace("default"))

	assert.Nil(t, err, "error listing left bindings")
	assert.Equal(t, 1, len(left.Items), "incorrent number of bindings left")
}
