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

package constraint

import (
	"context"
	"fmt"

	cpv1 "github.com/ciena/turnbuckle/pkg/apis/constraint/v1alpha1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ConstraintPolicyReconciler reconciles a Policy object.
// nolint:revive
type ConstraintPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// nolint:lll
//+kubebuilder:rbac:groups=constraint.ciena.com,resources=constraintpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=constraint.ciena.com,resources=constraintpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=constraint.ciena.com,resources=constraintpolicies/finalizers,verbs=update

// Reconcile evaluates updates to the requested constraint policy and updates
// internal status if required.
func (r *ConstraintPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.WithValues("constraintpolicy", req.NamespacedName)

	// fetch the policy in question
	var policy cpv1.ConstraintPolicy
	if err := r.Client.Get(context.TODO(), req.NamespacedName, &policy); err != nil {
		if kerrors.IsNotFound(err) || kerrors.IsGone(err) {
			logger.V(1).Info("not-found", lkNamespace, req.NamespacedName.Namespace,
				lkName, req.NamespacedName.Name)

			return ctrl.Result{}, nil
		}

		logger.V(1).Info(apiError,
			"error", err.Error(),
			"op", "get",
			lkNamespace, req.NamespacedName.Namespace,
			lkName, req.NamespacedName.Name)

		// nolint:wrapcheck
		return ctrl.Result{}, err
	}

	// Update the table subresource. this resource is simply to aid in the
	// display of the resource from the command line.
	list := []string{}

	for _, rule := range policy.Spec.Rules {
		list = append(list, rule.Name)
	}

	logger.V(1).Info("rule-list", "list", fmt.Sprintf("%+#v", list))

	if !areStringSlicesEqual(policy.Status.Table.Rules, list) {
		policy.Status.Table.Rules = list
		if err := r.Client.Status().Update(context.TODO(), &policy); err != nil {
			logger.V(1).Info("status-update-error", "error", err.Error())

			// nolint:wrapcheck
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConstraintPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&cpv1.ConstraintPolicy{}).
		Complete(r)
}
