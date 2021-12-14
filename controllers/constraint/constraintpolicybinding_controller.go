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
	"sort"
	"time"

	cpv1 "github.com/ciena/turnbuckle/apis/constraint/v1alpha1"
	"github.com/ciena/turnbuckle/apis/ruleprovider"
	pb "github.com/ciena/turnbuckle/pkg/pb"
	"github.com/ciena/turnbuckle/pkg/types"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	providerLabel           = "constraint.ciena.com/provider-%s"
	defaultEvaluationPeriod = 1 * time.Minute
	defaultRPCTimeout       = 10 * time.Second
)

// ConstraintPolicyBindingReconciler reconciles a Binding object.
// nolint:revive
type ConstraintPolicyBindingReconciler struct {
	client.Client
	Scheme                  *runtime.Scheme
	EvaluationErrorInterval time.Duration
	EvaluationInterval      time.Duration
	RPCTimeout              time.Duration
}

func detailsAreDifferent(left, right []*cpv1.ConstraintPolicyBindingStatusDetail) bool {
	if len(left) != len(right) {
		return true
	}

	// sort left and right based on policy name so we compare same to same
	sort.Slice(left, func(i, j int) bool {
		return left[i].Policy < left[j].Policy
	})
	sort.Slice(right, func(i, j int) bool {
		return right[i].Policy < right[j].Policy
	})

	for i := range left {
		if left[i].Policy != right[i].Policy ||
			left[i].Compliance != right[i].Compliance ||
			left[i].Reason != right[i].Reason {
			return true
		}

		// Sort rules by rule name for same to same comparison
		sort.Slice(left[i].RuleDetails, func(a, b int) bool {
			return left[i].RuleDetails[a].Rule < left[i].RuleDetails[b].Rule
		})
		sort.Slice(right[i].RuleDetails, func(a, b int) bool {
			return right[i].RuleDetails[a].Rule < right[i].RuleDetails[b].Rule
		})

		for j := range left[i].RuleDetails {
			if left[i].RuleDetails[j] != right[i].RuleDetails[j] {
				return true
			}
		}
	}

	return false
}

// nolint:lll
func (r *ConstraintPolicyBindingReconciler) evaluateRule(binding *cpv1.ConstraintPolicyBinding, rule *cpv1.ConstraintPolicyRule, svc *corev1.Service) (string, string) {
	dnsName := fmt.Sprintf("%s.%s.svc.cluster.local:5309", svc.Name, svc.Namespace)

	conn, err := grpc.Dial(dnsName, grpc.WithInsecure())
	if err != nil {
		return types.ComplianceError, err.Error()
	}

	// nolint:errcheck
	defer conn.Close()

	timeout := r.RPCTimeout
	if timeout == 0 {
		timeout = defaultRPCTimeout
	}

	client := ruleprovider.NewRuleProviderClient(conn)

	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	greq := ruleprovider.EvaluateRequest{
		Targets: pb.AsPBTarget(binding.Spec.Targets),
		Rule: &ruleprovider.PolicyRule{
			Name:    rule.Name,
			Request: rule.Request,
			Limit:   rule.Limit,
		},
	}

	gresp, err := client.Evaluate(ctx, &greq)
	if err != nil {
		return types.ComplianceError, err.Error()
	}

	return gresp.Compliance.String(), gresp.Reason
}

func (r *ConstraintPolicyBindingReconciler) checkAndUpdateStatus(
	binding *cpv1.ConstraintPolicyBinding,
	compliance, reason string,
	details []*cpv1.ConstraintPolicyBindingStatusDetail) error {
	// Only update status if something has changed
	changed := false

	if detailsAreDifferent(binding.Status.Details, details) {
		binding.Status.Details = details
		changed = true
	}

	if binding.Status.Compliance != compliance {
		binding.Status.LastComplianceChangeTimestamp = metav1.Now()
		binding.Status.Compliance = compliance
		changed = true
	}

	if binding.Status.FirstReason != reason {
		binding.Status.LastComplianceChangeTimestamp = metav1.Now()
		binding.Status.FirstReason = reason
		changed = true
	}

	// If the compliance is not not in violation, then clear any mitigation
	// timestamp
	if compliance != types.ComplianceViolation && !binding.Status.LastMitigatedTimestamp.IsZero() {
		binding.Status.LastMitigatedTimestamp = metav1.Time{Time: time.Time{}}
		changed = true
	}

	if changed {
		// nolint:wrapcheck
		return r.Client.Status().Update(context.TODO(), binding)
	}

	return nil
}

// nolint:lll
//+kubebuilder:rbac:groups=constraint.ciena.com,resources=constraintpolicybindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=constraint.ciena.com,resources=constraintpolicybindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=constraint.ciena.com,resources=constraintpolicybindings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// nolint:gocognit,funlen,cyclop
func (r *ConstraintPolicyBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("constraintpolicybinding",
		req.NamespacedName)

	// lookup binding in question
	var binding cpv1.ConstraintPolicyBinding
	if err := r.Client.Get(context.TODO(), req.NamespacedName, &binding); err != nil {
		if isNotFoundOrGone(err) {
			// if gone, no need to do anything
			logger.V(1).Info("not-found", lkNamespace, req.Namespace,
				lkName, req.Name)

			return ctrl.Result{}, nil
		}

		// retry on api-error
		logger.V(0).Info(apiError, "error", err.Error())

		// noling:nilerr
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	// Initialize the rule evaluation details and compliance. Set
	// initial compliance as Compliant as it is the lease severe level
	// and thus the status will be updated if a more severe value is
	// encountered
	// nolint: prealloc
	var details []*cpv1.ConstraintPolicyBindingStatusDetail

	var reason string

	compliance := types.ComplianceCompliant

	// Get the offer associated with this binding
	var offer cpv1.ConstraintPolicyOffer
	if err := r.Client.Get(context.TODO(), ktypes.NamespacedName{
		Namespace: req.Namespace, Name: binding.Spec.Offer,
	}, &offer); err != nil {
		if isNotFoundOrGone(err) {
			logger.V(0).Error(err, "offer-not-found", lkNamespace, req.Namespace,
				lkName, binding.Spec.Offer)

			// nolint:wrapcheck
			return ctrl.Result{}, err
		}

		logger.V(0).Error(err, apiError)

		// noling:nilerr
		return ctrl.Result{RequeueAfter: r.EvaluationErrorInterval}, nil
	}

	// Iterate over offers associated with the binding and collect the policies
	for _, pname := range offer.Spec.Policies {
		var policy cpv1.ConstraintPolicy
		if err := r.Client.Get(context.TODO(), ktypes.NamespacedName{
			Namespace: req.Namespace, Name: string(pname),
		}, &policy); err != nil {
			logger.V(0).Error(err, "policy-lookup-failed",
				lkNamespace, req.Namespace, lkName, pname)

			details = append(details, &cpv1.ConstraintPolicyBindingStatusDetail{
				Policy:     string(pname),
				Compliance: types.ComplianceError,
				Reason:     err.Error(),
			})

			continue
		}

		// If we have a policy, then set the the summary value to compliant,
		// to be overridden if a more severe evaluation is seen.
		detail := cpv1.ConstraintPolicyBindingStatusDetail{
			Policy:     string(pname),
			Compliance: types.ComplianceCompliant,
		}

		// Iterate over rules and evaluate
		for _, rule := range policy.Spec.Rules {
			logger.V(1).Info("rule-evaluation", "policy", pname, "rule", rule.Name)

			ruleDetail := cpv1.ConstraintPolicyBindingStatusRuleDetail{
				Rule: rule.Name,
			}

			var svcs corev1.ServiceList
			// nolint:gocritic
			if err := r.Client.List(context.TODO(), &svcs,
				client.InNamespace(req.Namespace),
				client.HasLabels([]string{fmt.Sprintf(providerLabel, rule.Name)})); err != nil {
				ruleDetail.Compliance = types.ComplianceError
				ruleDetail.Reason = err.Error()
			} else if len(svcs.Items) == 0 {
				ruleDetail.Compliance = types.ComplianceError
				ruleDetail.Reason = "rule provider not found"
			} else {
				// At least one provider was found, use the first one in
				// the list, but warn if more than a single one exists
				if len(svcs.Items) > 1 {
					logger.V(0).Info("multiple providers exists for rule", "rule",
						rule.Name)
				}
				ruleDetail.Compliance, ruleDetail.Reason = r.evaluateRule(&binding, rule, &(svcs.Items[0]))
			}

			detail.RuleDetails = append(detail.RuleDetails, &ruleDetail)

			logger.V(1).Info("rule evaluation summary",
				"binding", binding.Name,
				"policy", policy.Name,
				"rule", rule.Name,
				"current-summary-compliance", detail.Compliance,
				"rule-compliance", ruleDetail.Compliance)

			if types.CompareComplianceSeverity(detail.Compliance, ruleDetail.Compliance) > 0 {
				detail.Compliance = ruleDetail.Compliance
				detail.Reason = ruleDetail.Reason
			}
		}

		details = append(details, &detail)

		logger.V(1).Info("binding evalation summary",
			"binding", binding.Name,
			"current-compliance", compliance,
			"detail-compliance", detail.Compliance)

		if types.CompareComplianceSeverity(compliance, detail.Compliance) > 0 {
			compliance = detail.Compliance
			reason = detail.Reason
		}
	}

	// There is an optional evaluation period that can be specified as
	// part of the offer. If this optional value has been set, and can be
	// parsed as a duration, use it, else use the default
	period := r.EvaluationInterval

	if offer.Spec.Period != "" {
		var err error
		if period, err = time.ParseDuration(offer.Spec.Period); err != nil {
			logger.V(0).Error(err, "unable to parse period as duration",
				"period", offer.Spec.Period)

			period = r.EvaluationInterval
		}
	}

	if period == 0 {
		period = defaultEvaluationPeriod
	}

	logger.V(1).Info("requeue binding evaluation", lkNamespace, binding.Namespace,
		lkName, binding.Name, "period", period.String())

	return ctrl.Result{RequeueAfter: period}, r.checkAndUpdateStatus(&binding,
		compliance, reason, details)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConstraintPolicyBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&cpv1.ConstraintPolicyBinding{}).
		Complete(r)
}
