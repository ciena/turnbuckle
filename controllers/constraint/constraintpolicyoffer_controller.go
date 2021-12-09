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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	uv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	options "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	labelRef = "constraint.ciena.io/constraintPolicyOffer"
)

// ConstraintPolicyOfferReconciler reconciles a ConstraintPolicyOffer object
type ConstraintPolicyOfferReconciler struct {
	client.Client
	Scheme                       *runtime.Scheme
	OfferEvaluationErrorInterval time.Duration
	OfferEvaluationInterval      time.Duration
}

// visitState used to track if binding are still valid
type visitState struct {
	Visited bool
	Binding *cpv1.ConstraintPolicyBinding
}

func (r *ConstraintPolicyOfferReconciler) checkAndUpdateStatus(offer *cpv1.ConstraintPolicyOffer, count, compliant int) error {
	if offer.Status.BindingCount != count ||
		offer.Status.CompliantBindingCount != compliant ||
		offer.Status.BindingSelector != labelRef+"="+offer.Name {

		offer.Status.BindingCount = count
		offer.Status.CompliantBindingCount = compliant
		offer.Status.BindingSelector = labelRef + "=" + offer.Name

		return r.Client.Status().Update(context.TODO(), offer)
	}

	return nil
}

//+kubebuilder:rbac:groups=constraint.ciena.com,resources=constraintpolicyoffers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=constraint.ciena.com,resources=constraintpolicyoffers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=constraint.ciena.com,resources=constraintpolicyoffers/finalizers,verbs=update

func (r *ConstraintPolicyOfferReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("constraintpolicyoffer",
		req.NamespacedName)

	// Lookup the offer being reconciled
	var offer cpv1.ConstraintPolicyOffer
	if err := r.Client.Get(context.TODO(), req.NamespacedName, &offer); err != nil {

		if isNotFoundOrGone(err) {
			if r.Client.DeleteAllOf(context.TODO(), &cpv1.ConstraintPolicyBinding{},
				options.InNamespace(req.NamespacedName.Namespace),
				options.MatchingLabels(map[string]string{
					labelRef: req.NamespacedName.Name})) != nil {
				logger.V(0).Info("api-error", "error", err.Error())
				return ctrl.Result{RequeueAfter: r.OfferEvaluationErrorInterval}, nil
			}

			// Offer if gone, no need to retry
			return ctrl.Result{}, nil
		} else {
			logger.V(0).Info("api-error", "error", err.Error())
			return ctrl.Result{RequeueAfter: r.OfferEvaluationErrorInterval}, nil
		}
	}

	// Lookup all the bindings owned by this offer
	var bindings cpv1.ConstraintPolicyBindingList
	if err := r.Client.List(context.TODO(), &bindings,
		options.InNamespace(req.Namespace),
		options.MatchingLabels(map[string]string{
			labelRef: req.NamespacedName.Name})); err != nil {
		if !isNotFoundOrGone(err) {
			logger.V(0).Info("api-error", "error", err.Error())
			return ctrl.Result{RequeueAfter: r.OfferEvaluationErrorInterval}, nil
		}
	}

	// Create a visted map so we can track which bindings continue
	// to be neededgand which can be deleted.
	visits := make(map[string]*visitState)
	for i, binding := range bindings.Items {
		visits[binding.Name] = &visitState{
			Visited: false,
			Binding: &bindings.Items[i],
		}
	}

	refs := make(types.ReferencesMap)

	for _, target := range offer.Spec.Targets {
		targetName := target.Name
		if targetName == "" {
			targetName = "default"
		}
		// Add the empty target reference. This is important so that
		// when calculating permutations we know if we have have an empty
		// target list, because an empty target list means that there
		// are no permutations.
		refs[targetName] = types.References{}

		set, err := metav1.LabelSelectorAsMap(target.LabelSelector)
		if err != nil {
			logger.V(0).Error(err, "selector-parse-error", "namespace", req.NamespacedName.Namespace,
				"name", req.NamespacedName.Name, "type", targetName, "selector", target.LabelSelector)
			return ctrl.Result{RequeueAfter: r.OfferEvaluationErrorInterval}, nil
		}

		found := uv1.UnstructuredList{}
		found.SetKind(target.Kind)
		found.SetAPIVersion(target.APIVersion)
		if err := r.Client.List(context.TODO(), &found,
			options.InNamespace(req.NamespacedName.Namespace),
			options.MatchingLabels(set)); err != nil {

			if !isNotFoundOrGone(err) {
				logger.V(0).Error(err, "api-error")
				return ctrl.Result{RequeueAfter: r.OfferEvaluationErrorInterval}, nil
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
	for _, permutation := range refs.Permutations() {
		bindingName := permutation.AsBindingName(offer.ObjectMeta.Name)
		if visit, ok := visits[bindingName]; ok {
			visit.Visited = true
			continue
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
				Offer: offer.ObjectMeta.Name,
			},
			Status: cpv1.ConstraintPolicyBindingStatus{
				Compliance: "Pending",
				Details:    []cpv1.ConstraintPolicyBindingStatusDetail{},
			},
		}
		logger.V(0).Info("CREATE", "name", bindingName)
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
			if !isNotFoundOrGone(err) {
				logger.V(0).Error(err, "delete-binding", "name", name)
			}
		}
	}

	if err := r.checkAndUpdateStatus(&offer, count, compliant); err != nil {
		logger.V(0).Error(err, "status-update", "name", offer.ObjectMeta.Name)
		return ctrl.Result{RequeueAfter: r.OfferEvaluationErrorInterval}, nil
	}

	return ctrl.Result{RequeueAfter: r.OfferEvaluationInterval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConstraintPolicyOfferReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cpv1.ConstraintPolicyOffer{}).
		Complete(r)
}
