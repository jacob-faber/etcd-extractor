module github.com/camabeh/etcd-extractor

go 1.17

// Values from https://github.com/openshift/origin@release-4.8
require (
	github.com/openshift/api v0.0.0-20210521075222-e273a339932a
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	k8s.io/apimachinery v0.21.1
	k8s.io/kubectl v0.21.1
)

require (
	github.com/spf13/cobra v1.3.0
	go.etcd.io/bbolt v1.3.6 // indirect
	go.uber.org/zap v1.20.0
)

require (
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	golang.org/x/sys v0.0.0-20211205182925-97ca703d548d // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa // indirect
	google.golang.org/grpc v1.42.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.21.1 // indirect
	k8s.io/client-go v0.21.1 // indirect
	k8s.io/klog/v2 v2.8.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.0 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

// Fix for error:
//# go.etcd.io/etcd/clientv3/balancer/picker
//vendor/go.etcd.io/etcd/clientv3/balancer/picker/err.go:25:9: cannot use &errPicker{...} (type *errPicker) as type Picker in return argument:
//        *errPicker does not implement Picker (wrong type for Pick method)
//                have Pick(context.Context, balancer.PickInfo) (balancer.SubConn, func(balancer.DoneInfo), error)
//                want Pick(balancer.PickInfo) (balancer.PickResult, error)
//vendor/go.etcd.io/etcd/clientv3/balancer/picker/roundrobin_balanced.go:33:9: cannot use &rrBalanced{...} (type *rrBalanced) as type Picker in return argument:
//        *rrBalanced does not implement Picker (wrong type for Pick method)
//                have Pick(context.Context, balancer.PickInfo) (balancer.SubConn, func(balancer.DoneInfo), error)
//                want Pick(balancer.PickInfo) (balancer.PickResult, error)
//
// Newer grpc package changed Picker interface
replace google.golang.org/grpc => google.golang.org/grpc v1.27.1
