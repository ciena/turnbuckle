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
	"time"

	cpv1 "github.com/ciena/turnbuckle/apis/constraint/v1alpha1"
	"github.com/ciena/turnbuckle/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	uv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	labelRef = "constraint.ciena.io/constraintPolicyOffer"
)

// ConstraintPolicyOfferReconciler reconciles a ConstraintPolicyOffer object.
// nolint:revive
type ConstraintPolicyOfferReconciler struct {
	client.Client
	Scheme                  *runtime.Scheme
	EvaluationErrorInterval time.Duration
	EvaluationInterval      time.Duration
}

// visitState used to track if binding are still valid.
type visitState struct {
	Visited bool
	Binding *cpv1.ConstraintPolicyBinding
}

func (r *ConstraintPolicyOfferReconciler) checkAndUpdateStatus(
	offer *cpv1.ConstraintPolicyOffer, count, compliant int) error {
	if offer.Status.BindingCount != count ||
		offer.Status.CompliantBindingCount != compliant ||
		offer.Status.BindingSelector != labelRef+"="+offer.Name {
		offer.Status.BindingCount = count
		offer.Status.CompliantBindingCount = compliant
		offer.Status.BindingSelector = labelRef + "=" + offer.Name

		// nolint:wrapcheck
		return r.Client.Status().Update(context.TODO(), offer)
	}

	return nil
}

// nolint:lll
//+kubebuilder:rbac:groups=constraint.ciena.com,resources=constraintpolicyoffers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=constraint.ciena.com,resources=constraintpolicyoffers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=constraint.ciena.com,resources=constraintpolicyoffers/finalizers,verbs=update

// Reconcile reconciles updates to the offer resource instances.
// nolint:funlen,gocognit,cyclop
func (r *ConstraintPolicyOfferReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("constraintpolicyoffer",
		req.NamespacedName)

	// Lookup the offer being reconciled
	var offer cpv1.ConstraintPolicyOffer
	if err := r.Client.Get(context.TODO(), req.NamespacedName, &offer); err != nil {
		if kerrors.IsNotFound(err) || kerrors.IsGone(err) {
			if r.Client.DeleteAllOf(context.TODO(), &cpv1.ConstraintPolicyBinding{},
				client.InNamespace(req.NamespacedName.Namespace),
				client.MatchingLabels(map[string]string{
					labelRef: req.NamespacedName.Name,
				})) != nil {
				logger.V(0).Info(apiError, "error", err.Error())

				// nolint:nilerr
				return ctrl.Result{RequeueAfter: r.EvaluationErrorInterval}, nil
			}

			// Offer if gone, no need to retry
			return ctrl.Result{}, nil
		}

		logger.V(0).Info(apiError, "error", err.Error())

		return ctrl.Result{RequeueAfter: r.EvaluationErrorInterval}, nil
	}

	// Lookup all the bindings owned by this offer
	var bindings cpv1.ConstraintPolicyBindingList
	if err := r.Client.List(context.TODO(), &bindings,
		client.InNamespace(req.Namespace),
		client.MatchingLabels(map[string]string{
			labelRef: req.NamespacedName.Name,
		})); err != nil {
		if !kerrors.IsNotFound(err) || kerrors.IsGone(err) {
			logger.V(0).Info(apiError, "error", err.Error())

			return ctrl.Result{RequeueAfter: r.EvaluationErrorInterval}, nil
		}
	}

	// Create a visted map so we can track which bindings continue
	// to be neededgand which can be deleted.
	visits := make(map[string]*visitState)
	for i, binding := range bindings.Items {
		visits[binding.Name] = &visitState{
			Visited: false,
			Binding: bindings.Items[i],
		}
	}

	refs := make(types.ReferenceListMap)

	for _, target := range offer.Spec.Targets {
		targetName := target.Name
		if targetName == "" {
			targetName = "default"
		}

		// Add the empty target reference. This is important so that
		// when calculating permutations we know if we have have an empty
		// target list, because an empty target list means that there
		// are no permutations.
		refs[targetName] = types.ReferenceList{}

		set, err := metav1.LabelSelectorAsMap(target.LabelSelector)
		if err != nil {
			logger.V(0).Error(err, "selector-parse-error", lkNamespace, req.NamespacedName.Namespace,
				lkName, req.NamespacedName.Name, "type", targetName, "selector", target.LabelSelector)

			return ctrl.Result{RequeueAfter: r.EvaluationErrorInterval}, nil
		}

		found := uv1.UnstructuredList{}
		found.SetKind(target.Kind)
		found.SetAPIVersion(target.APIVersion)

		if err := r.Client.List(context.TODO(), &found,
			client.InNamespace(req.NamespacedName.Namespace),
			client.MatchingLabels(set)); err != nil {
			if !kerrors.IsNotFound(err) || kerrors.IsGone(err) {
				logger.V(0).Error(err, apiError)

				return ctrl.Result{RequeueAfter: r.EvaluationErrorInterval}, nil
			}
		}

		for _, item := range found.Items {
			ref := types.NewReferenceFromUnstructured(item)
			if !refs[targetName].Contains(ref) {
				refs[targetName] = append(refs[targetName], ref)
			}
		}
	}

	// There is an assumption that a binding is a unique combination of
	// instances, one from each target name. Therefore if any target
	// name resolves to no references all binding associated with the
	// offer are not valid and should be deleted.

	// Iterate over all tuples of targets
	keys, permutations := refs.Permutations()
	for _, permutation := range permutations {
		bindingName := permutation.AsBindingName(offer.ObjectMeta.Name)
		if visit, ok := visits[bindingName]; ok {
			visit.Visited = true

			continue
		}

		targets := make(map[string]types.Reference)

		for i := 0; i < len(keys); i++ {
			targets[keys[i]] = types.Reference{
				Cluster:    permutation[i].Cluster,
				APIVersion: permutation[i].APIVersion,
				Kind:       permutation[i].Kind,
				Namespace:  permutation[i].Namespace,
				Name:       permutation[i].Name,
			}
		}

		binding := cpv1.ConstraintPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: req.NamespacedName.Namespace,
				Name:      bindingName,
				Labels: map[string]string{
					labelRef: req.NamespacedName.Name,
				},
			},
			Spec: cpv1.ConstraintPolicyBindingSpec{
				Offer:   offer.ObjectMeta.Name,
				Targets: targets,
			},
			Status: cpv1.ConstraintPolicyBindingStatus{
				Compliance: "Pending",
				Details:    []*cpv1.ConstraintPolicyBindingStatusDetail{},
			},
		}

		logger.V(0).Info("CREATE", lkName, bindingName)

		if err := r.Client.Create(context.TODO(), &binding); err != nil {
			logger.V(0).Error(err, "binding-creation")

			continue
		}

		visits[bindingName] = &visitState{
			Visited: true,
			Binding: &binding,
		}
	}

	count := 0
	compliant := 0

	for name, visit := range visits {
		if visit.Visited {
			count++

			if visit.Binding.Status.Compliance == "Compliant" {
				compliant++
			}

			continue
		}

		if err := r.Client.Delete(context.TODO(), visit.Binding); err != nil {
			if !kerrors.IsNotFound(err) || kerrors.IsGone(err) {
				logger.V(0).Error(err, "delete-binding", lkName, name)
			}
		}
	}

	if err := r.checkAndUpdateStatus(&offer, count, compliant); err != nil {
		logger.V(0).Error(err, "status-update", lkName, offer.ObjectMeta.Name)

		return ctrl.Result{RequeueAfter: r.EvaluationErrorInterval}, nil
	}

	return ctrl.Result{RequeueAfter: r.EvaluationInterval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConstraintPolicyOfferReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&cpv1.ConstraintPolicyOffer{}).
		Complete(r)
}
