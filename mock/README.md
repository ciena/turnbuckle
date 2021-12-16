# Constraint Policy Rule Provider Mock

This directory contains Kubernetes custom resource and its controller that
can be used to mock rule providers for testing the constraint policy
Kubernetes capability.

## TL;DR

```
$ make docker-build docker-push install deploy
```

Then apply any required RuleProvider resources. Examples can be found
[here](./examples).

## What is a rule provider?

Within the Kubernetes constraint policy capability (turnbuckle) a rule
provider is a service that is used to evaluate a constraint rule against
a set of target resources and return the compliance of the rule. Specifically
a rule provider must implement the gRPC interface defined
[here](../apis/ruleprovider.proto).

## What is the mock controller

The mock controller implements the rule provider gRPC interface and uses
the configuration of the RuleProvider resources to determine how to 
respond to the `Evalaute` method in that interface.

The RuleProvider resources specify the following information:

- `rule` - the name of the rule to which the resource pertains.
- `targets` - the list of targets to which the resource pertains, specified
   as regex expressions (see below).
- `value` - the compliance value that should be returned if this resource
   matches the request.
- `reason` - the reason that is part of the gRPC response when the compliance
   level is not `Compliant`.
- `priority` - the priority of this resource instance (see below).

### Targets

The targets specified as part of the resource are used to match against the
targets specified as part of the gRPC request from the constraint policy
controller. Each target is specified as a tuple of `cluster-id`, `namespace`,
`apiVersion`, `kind`, and `name` separated by a colon (`:`).

_example_: `:default:v1/Pod/my-pod-79d849f576-kxhpl`

Note: in the example above the `cluster-id` was blank as this is left open
for future development for multi-cluster environments.

The target specification is evaluated as a _Python_ regular expression
against the request.

For a RuleProvider resource to be considered a match against a gRPC request
the name of the target and the target reference must match.

### Priority

The priority field in the RuleProvider resource is used to sort and
disambiguate the order in which RuleProvider resources are matched against a
request. When evaluating RuleProvider resources the list is iterated from
highest to lowest priority and the first match is accepted. When one or 
more RuleProvider resource has the same priority it is indeterminate 
which order the resouces s evalauted.

## Configuration

There is very litle configuration for this mock beside the individual 
RuleProvider resources instances.

### Default compliance level and reason

The default compliance level and reason returned when no RuleProvider resource
is found that matches a gRPC request can be set via the environment variables
`DEFAULT_COMPLIANCE` and `DEFAULT_REASON` when starting the controller. If
these values are not set then the values `Compliant` and `no value configured`
are returned.

### Provider indication

The constraint policy capability locates the rule provider implementation by
performing a search for service with a label of the form
`constraint.ciena.com/provider-<name>`, where `<name>` is the name of 
the constaint rule it supports. A service can support any number of constraint
rules.

To use this mock to test the constraint policy it is important to configure 
the service that deploys as part of this mock to support the rule you are
attempting to test against. Examples of labels added to the service can
be found in the kustomization patch
[here](./manifests/deploy/patch-service.yaml).
