package nilness

import (
	"time"

	"k8s.io/klog/v2"
)

func adversarialReturningExitHook(p *NamespacedName) {
	old := klog.OsExit
	klog.OsExit = func(int) {}
	defer func() { klog.OsExit = old }()
	if p == nil {
		klog.FlushAndExit(time.Nanosecond, 1)
	}
	klog.InfoS("reachable nil", "p", p) // want `logging value may panic when nil`
}

func adversarialReturningFatalHook(p *NamespacedName) {
	old := klog.OsExit
	klog.OsExit = func(int) {}
	defer func() { klog.OsExit = old }()
	if p == nil {
		klog.Fatal("nil")
	}
	klog.InfoS("returning fatal", "p", p) // want `logging value may panic when nil`
}

func adversarialReturningKlogExit(p *NamespacedName) {
	old := klog.OsExit
	klog.OsExit = func(int) { panic("exit intercepted") }
	defer func() { klog.OsExit = old }()
	if p == nil {
		klog.FlushAndExit(time.Nanosecond, 1)
	}
	klog.InfoS("normal continuation", "p", p)
}
