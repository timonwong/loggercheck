package nilness_test

import "k8s.io/klog/v2"

type externalValue struct{}

func (externalValue) String() string { return "external" }

func externalNilPointer(p *externalValue) {
	klog.InfoS("external nil", "p", p) // want `logging value may panic when nil`
}

func externalGuard(p *externalValue) {
	if p != nil {
		klog.InfoS("external safe", "p", p)
	}
}
