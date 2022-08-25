package klogonly

import (
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
)

func logrIgnored() {
	err := fmt.Errorf("error")

	log := logr.Discard()
	log = log.WithValues("key")
	log.Info("message", "key1", "value1", "key2", "value2", "key3")
	log.Error(err, "message", "key1", "value1", "key2")
	log.Error(err, "message", "key1", "value1", "key2", "value2")
}

func ExampleKlog() {
	err := fmt.Errorf("error")

	klog.InfoS("abc", "key1", "value1")
	klog.InfoS("abc", "key1", "value1", "key2") // want `odd number of arguments passed as key-value pairs for logging`

	klog.ErrorS(err, "abc", "key1", "value1")
	klog.ErrorS(err, "abc", "key1", "value1", "key2") // want `odd number of arguments passed as key-value pairs for logging`

	klog.V(1).InfoS("message", "key1", "value1")
	klog.V(1).InfoS("message", "key1", "value1", "key2") // want `odd number of arguments passed as key-value pairs for logging`

	klog.V(2).InfoSDepth(1, "message", "key1", "value1", "key2", "value2")
	klog.V(2).InfoSDepth(1, "message", "key1", "value1", "key2", "value2", "key3") // want `odd number of arguments passed as key-value pairs for logging`

	klog.V(3).ErrorS(err, "message", "key1", "value1")
	klog.V(3).ErrorS(err, "message", "key1") // want `odd number of arguments passed as key-value pairs for logging`

	// klog/v2 can expose logr logger
	logger := klog.NewKlogr()
	logger.Info("message", "key1") // will not check since logr is disabled
}
