# [![turnbuckle](./assets/turnbuckle_64x64.png)](https://github.com/ciena/turnbuckle) Turnbuckle - Extensible constraint policy for Kubernetes

## Summary

Extensible constraint policy model that uses labels selection to bind defined
constraints to resources (deployments, pods, services, service meshes, etc.)
enabling cross workload and cross cluster scheduling based on constraint
relationships.

## Definition

A constraint policy is a set of defined constraint rules, where each
constraint rule specifies a constraint and the requested and limit values for
that constraint. Examples of constraints include:

- _Latency_ -- network level constraint used to indicate delay in
  communication between workloads.

- _Jitter_ -- network level constraint used to indicate variable in latency in
  communication between workloads.

- _Bandwidth_ -- network level constraint used to indicate bandwidth required
   between resources

- _Encryption_ -- encryption of network communication between resources

(As the project defines an extensible constraint model this list is not meant
to List is not meant to be comprehensive or imply implementation in this
project)

## Extensibility

It is the goal of this project to allow for extensibility with respect to
constraints policy so that the library of constraints can be extended by
individual deployments without having to release the core of the extension.

## Relationship of Proposal to Other Projects and Components

### Scheduler

This project provides and extensions to the scheduler to account for the
constraint policies. This may [optionally] include interaction with external
capabilities to modify physical infrastructure on which the platform is run
and / or by which components on the platform communicate.

### Descheduler (https://github.com/kubernetes-sigs/descheduler)

This project leverages the work in the _Descheduler_ SIG to manage if/when a
resource is to be evicted from a Node based on the violation of a  constraint
policy rule. This is currently provided as `patch` against a specific release
of the descheduler project.

### Multicluster (https://github.com/kubernetes-sigs/kubefed)

This project intends to integrate with the federated Kubernetes work being
defined in the Multicluster SIG. This integration is not yet defined and
work will continue in this area.

## Overview of Constraint Policy Resources

This project extends a standard Kubernetes deployment by introducing custom
resource definitions (CRDs) that are used to define constraint policies and
bind them to schedulable Kubernetes resources.

- `ConstaintPolicy` - represents a named set of constraint rules that can be
  associated to one or more schedulable Kubernetes resources.

- `ConstraintPolicyOffer` - used to select the set of schedulable Kubernetes
  resources to which a ConstraintPolicy is applied.

- `ConstraintPolicyBinding` - represents a concrete or realized binding of a
  a ConstraintPolicy to a schedulable Kubernetes resource as well as the
  compliance status of the scheduled resources to the policy.

### ConnectPolicy

#### Example 1 - Encryption Constraint

```yaml
apiVersion: constraint.ciena.com/v1alpha1
kind: ConstraintPolicy
metadata:
  name: secure
spec:
  rules:
    - constraint: encrypted
      request: true
    - constraint: encrypted-bit-count
      request: 2048
```

#### Example 2 - Bandwidth Constraint

```yaml
apiVersion: constraint.ciena.com/v1alpha1
kind: ConstraintPolicy
metadata:
  name: bandwidth-restriction
spec:
  rules:
    - constraint: bandwidth
      request: 1MiB
      limit: 2KiB
```

#### Example 3 - Latency / Jitter Constraint

```yaml
apiVersion: constraint.ciena.com/v1alpha1
kind: ConstraintPolicy
metadata:
  name: low-latency
spec:
  rules:
    - constraint: latency
      request: 5ms
      limit: 20ms
  period: 1m
  grace: 2m
```

### ConstraintPolicyOffer

#### Example 1

```yaml
apiVersion: constraint.ciena.com/v1alpha1
kind: ConstraintPolicyOffer
metadata:
  name: slow-and-low
spec:
  source:
    clusterSelector:
      clusterType: edge
    nodeSelector:
      fpga: true
    podSelector:
      app: firewall
  destination:
    source:
      clusterSelector:
        clusterType: metro
      app: fw-rule-server
  policies:
    - bandwidth-restriction
    - low-latency
    - secure
  period: 10m
  grace: 1h
  violationPolicy: Evict
```

#### Example 2

```yaml
apiVersion: constraint.ciena.com/v1alpha1
kind: ConstraintPolicyOffer
metadata:
  name: fast-and-low
spec:
  source:
    matchLabels:
      endpointID: 8374643A-D965-4F23-9089-C3F19A02D688
  destination:
    podSelector:
      app: vr-server
  policies:
    - low-latency
    - high-bandwidth
  period: 15s
  grace: 1m
  violationPolicy: Evict
```

### Period and Grace

The resource type definitions contain attributes referenced as `period` and
`grace`. These attributes refer to how often (`period`) the constraint policy
should be evaluated and how long (`grace`) that the  constraint policy
evaluation is outside the specified range before it is considered a violation.

These attributes will be leveraged to update the status attributes when
evaluating policies.

### ViolationPolicy

The `ViolationPolicy` attribute references how the _system_ should react when
a constraint policy is violated. The values include `Ignore`, `Mediate`, and
`Evict`.

- `Ignore` --  indicates that the system should take no action when a policy
  is violated.

- `Mediate` -- indicates that the system should attempt to bring the policy
  into compliance without disrupting a workload, if possible.

- `Evict` -- indicates the the system should evict the resource such that an
  attempt will be made to re-schedule the resource in such a way that the
  policy will be in compliance.

## Concrete Constraint Policy Bindings from Offers

A `ConstraintPolicyOffer` specifies a selector via which the subjects of of a
policy are identified. Because this selector may map to multiple Kubernetes
resources instances of a ConstraintPolicyOffer are concretely realized with
the creation of a `ConstraintPolicyBinding` for each subject pair.

The lifecycle of ConstraintPolicyBinding instances are managed by the
ConstraintPolicyOffer controller.

The status of the ConstraintPolicyBinding instances report the compliance
state of the concrete binding with its originating ConstraintPolicyOffer.

```bash
TODO - insert output of get constraintpolicybindings command
```

## Rule Extensibility

It is not possible to define all expected constraints that a deployment or
operator may consider. To enable extensibility without having to re-release or
have implementors "fork" the core of the capability the following is
implemented to support rule extensions.

A `PolicyRuleProvider` is be a pod that implements the logic for one or more
constraint rules, i.e. latency, jitter, etc. This pod implements a defined
gRPC interface that is utilized by the scheduler and evaluation capabilities
of this project.

```go
TODO - insert snippet of interface definition
```

The binding of a rule type to its implementation will be managed by applying
labels to the services which implement the provider interface for the metrics.
The labels are of a well know format:

```
constraint.ciena.com/constraint-<metric-name>
```

The advantage of providing an exposed mechanism to bind the provider to an
implementation is that it enables visibility of the binding to the operator as
well as it enables the operator to dynamically modify the binding including
the ability to upgrade a operator in a prescriptive manner.


## Implementation considerations

### Interaction with external resource managers

Still under development is an API to enable the management or external
resources, such as an underlay network, as part of the scheduling and
mediation work. The idea being that if constraints cannot be met or are
violated the external resources managers can be leveraged to influence
(change) the environment in such a way that the constraints might be met.
