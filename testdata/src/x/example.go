package x

import (
	"log"
	"os"

	"k8s.io/klog/v2"
)

type Value struct{}

func (Value) String() string { return "value" }

func adversarialVendoredLog(p *Value) {
	if p == nil {
		log.Fatal("nil")
	}
	klog.InfoS("returning vendored log", "p", p) // want `logging value may panic when nil`
}

func adversarialVendoredOS(p *Value) {
	if p == nil {
		os.Exit(1)
	}
	klog.InfoS("returning vendored os", "p", p) // want `logging value may panic when nil`
}
