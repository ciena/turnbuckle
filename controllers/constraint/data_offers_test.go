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
	cpv1 "github.com/ciena/turnbuckle/apis/constraint/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// nolint:gochecknoglobals
var offers = cpv1.ConstraintPolicyOfferList{
	TypeMeta: metav1.TypeMeta{
		Kind:       "ConstraintPolicyOffer",
		APIVersion: "v1alpha1",
	},
	ListMeta: metav1.ListMeta{},
	Items: []*cpv1.ConstraintPolicyOffer{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConstraintPolicyOffer",
				APIVersion: "v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "gold-offer",
			},
			Spec: cpv1.ConstraintPolicyOfferSpec{
				Targets: []*cpv1.ConstraintPolicyOfferTarget{
					{
						Name:       "source",
						APIVersion: "v1",
						Kind:       "Pod",
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "hello-client",
							},
						},
					},
					{
						Name:       "destination",
						APIVersion: "v1",
						Kind:       "Pod",
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "hello-server",
							},
						},
					},
				},
				Policies: []cpv1.Policy{
					"gold-policy",
				},
				ViolationPolicy: "Evict",
				Period:          "5s",
				Grace:           "1m",
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConstraintPolicyOffer",
				APIVersion: "v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "large-offer",
			},
			Spec: cpv1.ConstraintPolicyOfferSpec{
				Targets: []*cpv1.ConstraintPolicyOfferTarget{
					{
						APIVersion: "v1",
						Kind:       "Pod",
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "hello-server",
							},
						},
					},
				},
				Policies: []cpv1.Policy{
					"large",
				},
				ViolationPolicy: "Evict",
				Period:          "5s",
				Grace:           "1m",
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConstraintPolicyOffer",
				APIVersion: "v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "complex-offer",
			},
			Spec: cpv1.ConstraintPolicyOfferSpec{
				Targets: []*cpv1.ConstraintPolicyOfferTarget{
					{
						Name:       "one",
						APIVersion: "v1",
						Kind:       "Pod",
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "hello-client",
							},
						},
					},
					{
						Name:       "two",
						APIVersion: "v1",
						Kind:       "Pod",
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "hello-server",
							},
						},
					},
					{
						Name:       "three",
						APIVersion: "v1",
						Kind:       "Pod",
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "hello-monitor",
							},
						},
					},
				},
				Policies: []cpv1.Policy{
					"complex",
				},
				ViolationPolicy: "Evict",
				Period:          "5s",
				Grace:           "1m",
			},
		},
	},
}
