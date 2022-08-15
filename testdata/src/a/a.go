package a

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
)

func Example() {
	log := logr.Discard()
	log = log.WithValues("key")                                         // want `odd number of arguments passed as key-value pairs for logging`
	log.Info("message", "key1", "value1", "key2", "value2", "key3")     // want `odd number of arguments passed as key-value pairs for logging`
	log.Error(fmt.Errorf("error"), "message", "key1", "value1", "key2") // want `odd number of arguments passed as key-value pairs for logging`
	log.Error(fmt.Errorf("error"), "message", "key1", "value1", "key2", "value2")

	var log2 logr.Logger
	log2 = log
	log2.Info("message", "key1") // want `odd number of arguments passed as key-value pairs for logging`

	log3 := logr.FromContextOrDiscard(context.TODO())
	log3.Error(fmt.Errorf("error"), "message", "key1") // want `odd number of arguments passed as key-value pairs for logging`
}
