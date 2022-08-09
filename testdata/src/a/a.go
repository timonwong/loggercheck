package a

import (
	"fmt"

	"github.com/go-logr/logr"
)

func Example() {
	log := logr.Discard()
	log = log.WithValues("key")                                         // error
	log.Info("messsage", "key1", "value1", "key2", "value2", "key3")    // error
	log.Error(fmt.Errorf("error"), "message", "key1", "value1", "key2") // error
	log.Error(fmt.Errorf("error"), "message", "key1", "value1", "key2", "value2")
}
