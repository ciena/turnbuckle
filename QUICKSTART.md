# Quick start guide

The purpose of this document is to a short walk through of building, deploying,
and using the constraint policy resources as a way to help demonstrate the
capability.

# Local Kubernetes cluster

This quick start guide use Kubernetes in Docker (KinD) to create a Kubernetes
cluster for the walk through. If you choose to use your own cluster you will
have to adjust the commands to your environment.

Find instructions to install KinD at `https://kind.sigs.k8s.io/docs/user/quick-start/#installation`.

## Create Kubernetes cluster

The following script creates a Kubernetes cluster with a single control plane
node and 3 compute nodes. Additionally, this script creates a local docker
repository and attaches it to the Kubernetes cluster. This script it based
on a script on the KinD website: `https://kind.sigs.k8s.io/docs/user/local-registry/`.

```bash
#!/bin/sh
set -o errexit

# create registry container unless it already exists
reg_name='kind-registry'
reg_port='5000'
if [ "$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)" != 'true' ]; then
  docker run \
    -d --restart=always -p "127.0.0.1:${reg_port}:5000" --name "${reg_name}" \
    registry:2
fi

# create a cluster with the local registry enabled in containerd
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_name}:5000"]
EOF

# connect the registry to the cluster network if not already connected
if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${reg_name}")" = 'null' ]; then
  docker network connect "kind" "${reg_name}"
fi

# Document the local registry
# https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF
```

# Build and deploy

## Docker registry

Define the docker registry to use with your Kubernetes cluster

```bash
export DOCKER_REGISTRY=localhost:5000
```

## Build and deploy constraint policy components

The custom resource definitions (CRDs) are the definitions of the resource type
introduced by the constraint policy. The controllers implement the behviors for
those CRDs.

```bash
make docker-build docker-push install deploy
```

## Build and deploy the mock rule provider

The mock rule provider is a component that is used in development and testing,
but can also be used for demonstration purposes. This component is a set of 
CRDs that enables you to mock the responses for rule provider implementations.
In a production deployment you would not use this component as there would be
services in your Kubernetes environment that implement the rule providers.

```bash
make -C mock docker-build docker-push install deploy
```

## Build and deploy descheduler with patch for constraint policy

```bash
cd descheduler
git clone --branch v0.22.1 --depth 1 https://github.com/kubernetes-sigs/descheduler
cd descheduler
patch -s -p1  < ../descheduler-v0.22.1.patch
go mod verify
go mod tidy
go mod vendor
VERSION=cp make image
docker tag descheduler:cp $DOCKER_REGISTRY/descheduler:cp
docker push $DOCKER_REGISTRY/descheduler:cp
./deploy.sh -t cp -r localhost:5000/descheduler -k $HOME/.kube/config
```

## Checkpoint

At this point in the walkthough you should be able to see an output similar
to the following when querying Kubernetes for crds, services, and pods.

```bash
kubectl get --all-namespaces crd,svc,pod | grep -v kube-system
NAME                                                                                          CREATED AT
customresourcedefinition.apiextensions.k8s.io/constraintpolicies.constraint.ciena.com         2022-01-28T18:56:12Z
customresourcedefinition.apiextensions.k8s.io/constraintpolicybindings.constraint.ciena.com   2022-01-28T18:56:12Z
customresourcedefinition.apiextensions.k8s.io/constraintpolicyoffers.constraint.ciena.com     2022-01-28T18:56:12Z
customresourcedefinition.apiextensions.k8s.io/costproviders.constraint.ciena.com              2022-01-28T19:09:42Z
customresourcedefinition.apiextensions.k8s.io/ruleproviders.constraint.ciena.com              2022-01-28T19:09:42Z

NAMESPACE           NAME                                                    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                  AGE
default             service/kubernetes                                      ClusterIP   10.96.0.1      <none>        443/TCP                  93m
default             service/mock-provider                                   ClusterIP   10.96.35.210   <none>        5309/TCP                 73m
turnbuckle-system   service/turnbuckle-controller-manager-metrics-service   ClusterIP   10.96.211.30   <none>        8443/TCP                 86m

NAMESPACE            NAME                                                 READY   STATUS      RESTARTS   AGE
default              pod/mock-manager-5bf9f88cc-qx8f5                     1/1     Running     0          73m
local-path-storage   pod/local-path-provisioner-547f784dff-84rn2          1/1     Running     0          93m
turnbuckle-system    pod/constraint-policy-scheduler-55bc8bbb4c-rw26w     1/1     Running     0          86m
turnbuckle-system    pod/descheduler-27390022-9rf24                       0/1     Completed   0          46s
turnbuckle-system    pod/turnbuckle-controller-manager-6d6cdbc786-pfg6p   2/2     Running     0          86m
```

# Deploy initial constraint policies

The following command will create a `ConstraintPolicy` with a single rule,
named `latency` with a `request` and `limit` values. It also defines a
`ConstraintPolicyOffer` that binds the policy to a `hello-client` and a
`hello-server` via label selection.

```bash
kubectl apply -f examples/wt-policies.yaml
```

# Configure rule provider mock

The following command creates a configuration so that the mock rule provider
will return an infinite cost when evaluation nodes for scheduling as well as
a `Violation` response when evaluating the `latency` rule. This configuration
will prevent our sample Pods from being deployed because no nodes will be 
found that meet the constraints.

```bash
kubectl apply -f examples/wt-ruleprovider.yaml
```

# Deploy workloads

```bash
kubectl apply -f examples/wt-deploy.yaml
```

Because no nodes can be found that meet the constraints the `hello-server` and
`hello-client` pod will be stuct in `Pending` status.

```bash
$ kubectl get po
NAME                           READY   STATUS    RESTARTS   AGE
hello-client-69fd7d8c5-swzcl   0/1     Pending   0          10s
hello-server-74cfb4f9-n6tbt    0/1     Pending   0          10s
mock-manager-5bf9f88cc-qx8f5   1/1     Running   0          12m
```

# Modify mock to `Compliant`

Patch the `latency-rule` definition to return `Compliant`.

```bash
kubectl patch ruleprovider latency-rule --type 'merge' -p '{"spec":{"value":"Compliant","reason":"ok"}}'
```

Verify the rule is not returning `Compliant`.

```bash
$ kubectl get rp/latency-rule
NAME           RULE      PRIORITY   VALUE       AGE
latency-rule   latency   10         Compliant   15m
```

Pods will now be sccheduled to a node and created. This could take up to 100s
based on default configured retry periods.

```bash
$ kubectl get pods -o wide
NAME                           READY   STATUS    RESTARTS   AGE   IP           NODE           NOMINATED NODE   READINESS GATES
hello-client-69fd7d8c5-swzcl   1/1     Running   0          13m   10.244.3.3   kind-worker3   <none>           <none>
hello-server-74cfb4f9-n6tbt    1/1     Running   0          13m   10.244.1.3   kind-worker2   <none>           <none>
mock-manager-5bf9f88cc-qx8f5   1/1     Running   0          26m   10.244.2.3   kind-worker    <none>           <none>
```

Verify the `ConstraintPolicyBinding` status.

```bash
$ kubectl get cpb
NAME                    OFFER        STATUS      REASON   AGE
fast-offer-6f668f96c5   fast-offer   Compliant            9m56s
```

# Modify mock to `Limit`

This will patch the rule provider to return `Limit` as opposed to compliant.
This will not cause any workload to be evicted or rescheduled, but will
update the binding to the new value.

```bash
kubectl patch ruleprovider latency-rule --type 'merge' -p '{"spec":{"value":"Limit","reason":"could not meet request"}}'
```

Verify the rule is eturning `Limit`.

```bash
$ kubectl get rp/latency-rule
NAME           RULE      PRIORITY   VALUE   AGE
latency-rule   latency   10         Limit   24m
```

Verify the `ConstraintPolicyBinding` status.

```bash
$ kubectl get cpb
NAME                    OFFER        STATUS   REASON                   AGE
fast-offer-6f668f96c5   fast-offer   Limit    could not meet request   17m
```

# Modify mock to `Violation`

This will patch the rule provider to return `Violation` causing the descheduler
to eventually evict a pod after the configured `grace` period has expired.

```bash
kubectl patch ruleprovider latency-rule --type 'merge' -p '{"spec":{"value":"Violation","reason":"network capability not available"}}'
```

Verify the rule is eturning `Limit`.

```bash
$ kubectl get rp/latency-rule
NAME           RULE      PRIORITY   VALUE       AGE
latency-rule   latency   10         Violation   27m
```

Verify the `ConstraintPolicyBinding` status.

```bash
$ kubectl get cpb
NAME                    OFFER        STATUS      REASON                             AGE
fast-offer-6f668f96c5   fast-offer   Violation   network capability not available   20m
```

One of the "hello" pods should not be scheduled and be stuck in the `Pending`
status because the constraint cannot be met using any node.

```bash
$ kubectl get pods
NAME                           READY   STATUS    RESTARTS   AGE
hello-client-69fd7d8c5-swzcl   1/1     Running   1          52m
hello-server-74cfb4f9-42mrd    0/1     Pending   0          78s
mock-manager-5bf9f88cc-qx8f5   1/1     Running   0          65m
```

If you patch the rule provider back to `Compliant` or `Limit` the pending pod
will be scheduled to a node.

