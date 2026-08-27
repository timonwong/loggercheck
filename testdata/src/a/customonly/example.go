package customonly

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
)

type valueStringer struct{}

func (valueStringer) String() string { return "value" }

func logrIgnored() {
	err := fmt.Errorf("error")

	log := logr.Discard()
	log = log.WithValues("key")
	log.Info("message", "key1", "value1", "key2", "value2", "key3")
	log.Error(err, "message", "key1", "value1", "key2")
	log.Error(err, "message", "key1", "value1", "key2", "value2")
}

func ExampleCustomLogger() {
	err := errors.New("example error")

	var valueStringerPtr *valueStringer

	// custom SugaredLogger
	log := New()
	defer log.Sync()

	// won't bail because rules file wrong
	log.XXXDebugw("message", "key1")

	log.Infow("abc", "key1", "value1")
	log.Infow("abc", "key1", valueStringerPtr) // want `logging value may panic when nil because its element type implements fmt.Stringer`
	log.Infow("abc", "key1", "value1", "key2") // want `odd number of arguments passed as key-value pairs for logging`

	log.Errorw("message", "err", err, "key1", "value1")
	log.Errorw("message", err, "key1", "value1", "key2", "value2") // want `odd number of arguments passed as key-value pairs for logging`

	// with test
	log.With("with_key1", "with_value1").Infow("message", "key1", "value1")
	log.With("with_key1", "with_value1").Infow("message", "key1", "value1", "key2") // want `odd number of arguments passed as key-value pairs for logging`
	log.With("with_key1").Infow("message", "key1", "value1")                        // want `odd number of arguments passed as key-value pairs for logging`
}

func ExampleCustomLoggerPackageLevelFunc() {
	err := errors.New("example error")

	defer Sync()

	Infow("abc", "key1", "value1")
	Infow("abc", "key1", "value1", "key2") // want `odd number of arguments passed as key-value pairs for logging`

	Errorw("message", "err", err, "key1", "value1")
	Errorw("message", err, "key1", "value1", "key2", "value2") // want `odd number of arguments passed as key-value pairs for logging`

	// with test
	With("with_key1", "with_value1").Infow("message", "key1", "value1")
	With("with_key1", "with_value1").Infow("message", "key1", "value1", "key2") // want `odd number of arguments passed as key-value pairs for logging`
	With("with_key1").Infow("message", "key1", "value1")                        // want `odd number of arguments passed as key-value pairs for logging`
}
