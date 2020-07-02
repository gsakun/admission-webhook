module github.com/cnych/admission-webhook

go 1.13

require (
	github.com/ghodss/yaml v1.0.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/btree v1.0.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/rancher/norman v0.0.0-20191209163739-5b9227fe3222
	github.com/sirupsen/logrus v1.4.2
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.5.2 // indirect
)

replace (
	github.com/hd-Li/types => github.com/gsakun/types v0.0.0-20200612144151-977065b44fda
	k8s.io/api => k8s.io/api v0.0.0-20181004124137-fd83cbc87e76
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181004124836-1748dfb29e8a
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20180913025736-6dd46049f395
	k8s.io/client-go => k8s.io/client-go v9.0.0+incompatible
	k8s.io/kubernetes => k8s.io/kubernetes v1.12.2
)
