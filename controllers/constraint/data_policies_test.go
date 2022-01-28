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
	cpv1 "github.com/ciena/turnbuckle/pkg/apis/constraint/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// nolint:gochecknoglobals
var policies = cpv1.ConstraintPolicyList{
	TypeMeta: metav1.TypeMeta{
		Kind:       "ConstraintPolicy",
		APIVersion: "v1alpha1",
	},
	ListMeta: metav1.ListMeta{},
	Items: []cpv1.ConstraintPolicy{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConstraintPolicy",
				APIVersion: "v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "complex",
			},
			Spec: cpv1.ConstraintPolicySpec{
				Rules: []*cpv1.ConstraintPolicyRule{
					{
						Name:    "space",
						Request: "4G",
						Limit:   "2G",
					},
				},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConstraintPolicy",
				APIVersion: "v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "large",
			},
			Spec: cpv1.ConstraintPolicySpec{
				Rules: []*cpv1.ConstraintPolicyRule{
					{
						Name:    "cpu-count",
						Request: "4",
						Limit:   "2",
					},
					{
						Name:    "memory",
						Request: "8G",
						Limit:   "2G",
					},
				},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConstraintPolicy",
				APIVersion: "v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "gold-policy",
			},
			Spec: cpv1.ConstraintPolicySpec{
				Rules: []*cpv1.ConstraintPolicyRule{
					{
						Name:    "latency",
						Request: "200us",
						Limit:   "500us",
					},
					{
						Name:    "jitter",
						Request: "10ms",
						Limit:   "15ms",
					},
				},
			},
		},
	},
}
