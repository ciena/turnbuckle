module github.com/ciena/turnbuckle

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.2.0
	github.com/golang/protobuf v1.4.3
	github.com/julienschmidt/httprouter v1.2.0
	github.com/networkservicemesh/networkservicemesh/k8s v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/spf13/cobra v1.1.1 // indirect
	github.com/spf13/viper v1.7.0 // indirect
	github.com/tatsushid/go-fastping v0.0.0-20160109021039-d7bb493dee3e // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489 // indirect
	go.uber.org/zap v1.15.0
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb // indirect
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	google.golang.org/grpc v1.27.1
	google.golang.org/protobuf v1.25.0
	k8s.io/api v0.20.0
	k8s.io/apiextensions-apiserver v0.18.6 // indirect
	k8s.io/apimachinery v0.20.0
	k8s.io/apiserver v0.20.0
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.20.0
	k8s.io/kube-scheduler v0.18.6
	sigs.k8s.io/controller-runtime v0.5.0
	sigs.k8s.io/descheduler v0.10.0
)

replace (
	github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8
	github.com/networkservicemesh/networkservicemesh => github.com/networkservicemesh/networkservicemesh v0.0.0-20191017074247-aa5815869b2c
	github.com/networkservicemesh/networkservicemesh/controlplane => github.com/networkservicemesh/networkservicemesh/controlplane v0.0.0-20191017074247-aa5815869b2c
	github.com/networkservicemesh/networkservicemesh/controlplane/api => github.com/networkservicemesh/networkservicemesh/controlplane/api v0.0.0-20191017074247-aa5815869b2c
	github.com/networkservicemesh/networkservicemesh/dataplane => github.com/networkservicemesh/networkservicemesh/dataplane v0.0.0-20191017074247-aa5815869b2c
	github.com/networkservicemesh/networkservicemesh/dataplane/api => github.com/networkservicemesh/networkservicemesh/dataplane/api v0.0.0-20191017074247-aa5815869b2c
	github.com/networkservicemesh/networkservicemesh/k8s => github.com/networkservicemesh/networkservicemesh/k8s v0.0.0-20191017074247-aa5815869b2c
	github.com/networkservicemesh/networkservicemesh/k8s/api => github.com/networkservicemesh/networkservicemesh/k8s/api v0.0.0-20191017074247-aa5815869b2c
	github.com/networkservicemesh/networkservicemesh/pkg => github.com/networkservicemesh/networkservicemesh/pkg v0.0.0-20191017074247-aa5815869b2c
	github.com/networkservicemesh/networkservicemesh/sdk => github.com/networkservicemesh/networkservicemesh/sdk v0.0.0-20191017074247-aa5815869b2c
	github.com/networkservicemesh/networkservicemesh/utils => github.com/networkservicemesh/networkservicemesh/utils v0.0.0-20191017074247-aa5815869b2c
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191016112112-5190913f932d
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191016115326-20453efc2458
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191016115129-c07a134afb42
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191016111319-039242c015a9
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191016115521-756ffa5af0bd
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191016112429-9587704a8ad4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191016114939-2b2b218dc1df
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191016114407-2e83b6f20229
	//	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191016114748-65049c67a58b
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20200326032648-3033f8d0a303
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191016114556-7841ed97f1b2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191016115753-cf0698c3a16b
	k8s.io/metrics => k8s.io/metrics v0.0.0-20191016113814-3b1a734dba6e
	//k8s.io/metrics => k8s.io/metrics v0.0.0-20191016113814-3b1a734dba6e
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191016112829-06bb3c9d77c9
)
