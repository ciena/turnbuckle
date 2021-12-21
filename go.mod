module github.com/ciena/turnbuckle

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/golang/protobuf v1.5.2
	github.com/julienschmidt/httprouter v1.3.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	go.uber.org/tools v0.0.0-20190618225709-2cfd321de3ee // indirect
	go.uber.org/zap v1.19.0
	golang.org/x/net v0.0.0-20211209124913-491a49abca63
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1 // indirect
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/apiserver v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/component-base v0.22.2
	k8s.io/kube-scheduler v0.20.2
	sigs.k8s.io/controller-runtime v0.10.3
	sigs.k8s.io/descheduler v0.20.0
)
