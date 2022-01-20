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
imitations under the License.
*/

package constraint_test

import (
	cpv1 "github.com/ciena/turnbuckle/pkg/apis/constraint/v1alpha1"
	types "github.com/ciena/turnbuckle/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// nolint:gochecknoglobals
var bindings = cpv1.ConstraintPolicyBindingList{
	TypeMeta: metav1.TypeMeta{
		Kind:       "ConstraintPolicyBinding",
		APIVersion: "v1alpha1",
	},
	ListMeta: metav1.ListMeta{},
	Items: []*cpv1.ConstraintPolicyBinding{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConstraintPolicyBinding",
				APIVersion: "v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "gold-offer-binding",
				Labels: map[string]string{
					"constraint.ciena.io/constraintPolicyOffer": "gold-offer",
				},
			},
			Spec: cpv1.ConstraintPolicyBindingSpec{
				Targets: map[string]types.Reference{
					"source": {
						Cluster:    "",
						APIVersion: "v1",
						Kind:       "Pod",
						Namespace:  "default",
						Name:       "foo",
					},
					"destination": {
						Cluster:    "",
						APIVersion: "v1",
						Kind:       "Pod",
						Namespace:  "default",
						Name:       "bar",
					},
				},
			},
		},
	},
}
