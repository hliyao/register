module register

go 1.12

replace (
	cloud.google.com/go => github.com/googleapis/google-cloud-go v0.26.0
	go.etcd.io/bbolt => github.com/etcd-io/bbolt v1.3.2
	go.mongodb.org/mongo-driver => github.com/mongodb/mongo-go-driver v1.0.1
	go.uber.org/multierr => github.com/uber-go/multierr v1.1.0
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/exp => github.com/golang/exp v0.0.0-20190121172915-509febef88a4
	golang.org/x/lint => github.com/golang/lint v0.0.0-20181026193005-c67002cb31c3
	golang.org/x/net => github.com/golang/net v0.0.0-20190320064053-1272bf9dcd53
	golang.org/x/oauth2 => github.com/golang/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/sync => github.com/golang/sync v0.0.0-20181108010431-42b317875d0f
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190219203350-90b0e4468f99 // indirect
	golang.org/x/text => github.com/golang/text v0.3.0
	golang.org/x/time => github.com/golang/time v0.0.0-20190308202827-9d24e82272b4
	golang.org/x/tools => github.com/golang/tools v0.0.0-20180221164845-07fd8470d635
	google.golang.org/appengine => github.com/golang/appengine v1.1.0
	google.golang.org/genproto => github.com/google/go-genproto v0.0.0-20190307195333-5fe7a883aa19
	google.golang.org/grpc => github.com/grpc/grpc-go v1.19.0
	gopkg.in/alecthomas/kingpin.v2 => github.com/alecthomas/kingpin v2.2.6+incompatible
	gopkg.in/go-playground/assert.v1 => github.com/go-playground/assert v1.2.1 // indirect
	gopkg.in/go-playground/validator.v8 => github.com/go-playground/validator v8.18.2+incompatible // indirect
	gopkg.in/inf.v0 => github.com/go-inf/inf v0.9.1 // indirect
	gopkg.in/yaml.v2 => github.com/go-yaml/yaml v0.0.0-20181115110504-51d6538a90f8
	k8s.io/api => github.com/kubernetes/api v0.0.0-20180510142701-9273ee02527c
	k8s.io/apimachinery => github.com/kubernetes/apimachinery v0.0.0-20180925215425-1926e7bb5c13
	k8s.io/kube-openapi => github.com/kubernetes/kube-openapi v0.0.0-20190215190454-ea82251f3668 // indirect
)

require (
	github.com/astaxie/beego v1.12.0
	github.com/coreos/etcd v3.3.12+incompatible
	golang.org/x/net v0.0.0-20190320064053-1272bf9dcd53
	google.golang.org/grpc v1.19.0
	gopkg.in/yaml.v2 v2.2.1
)
