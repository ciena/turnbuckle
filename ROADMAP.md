# Turnbuckle Roadmap

This file list features or capabilities the the turnbuckle project plans
to incorporate in the future. While a specific timeline or release is not
specified they are listed in priority order. The contents of this roadmap
will change over time, both priority and context as discussions and work
continues.

Roadmap items are not considered formally under development or plan until
feature issues are created to represent the development.

## Constraint mediation controller

When scheduling a workload or mediating a violation of a constraint policy
the scheduler / descheduler can _reach out_ to a mediation controller
to attempt to adjust the underlying physical infrastructure (or other) to
correct any violation without requiring an eviction / rescheduler of the
workload to restore policy compliance.

## Multicluster support

The constraint policy was design to be incorporated into a multi-cluster
environment, as evidenced by the `cluster` reference in the model. This
roadmap item plans to add a multi-cluster scheduler component. This may
include integration with one of the multi-cluster projects such as `KubeFed`.
