package requirestringkey

import (
	"fmt"

	"github.com/go-logr/logr"
)

func ExampleRequireStringKey() {
	err := fmt.Errorf("error")

	log := logr.Discard()
	log.Error(err, "message", 1, "value1")              // want `logging keys must be of type string`
	log.Error(err, "message", []byte("key1"), "value1") // want `logging keys must be of type string`
	key1 := []byte("key1")
	log.Error(err, "message", key1, "value1") // want `logging keys must be of type string`
	log.Error(err, "message", string(key1), "value1")
	log.Error(err, "message", func() bool { return true }(), "value1") // want `logging keys must be of type string`

	type Str string
	key2 := Str("key2")
	log.Error(err, "message", func() string { return "key1" }(), "value1")
	log.Error(err, "message", "key1", "value1", key2, "value2") // want `logging keys must be of type string`

	type String = string
	log.Error(err, "message", String(key1), "value1")
}
