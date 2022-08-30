package requirestringkey

import (
	"fmt"

	"github.com/go-logr/logr"
)

func ExampleRequireStringKey() {
	err := fmt.Errorf("error")

	log := logr.Discard()
	log.Error(err, "message", 1, "value1")              // want `logging keys are expected to be inlined constant strings, please replace "1" provided with string`
	log.Error(err, "message", []byte("key1"), "value1") // want `logging keys are expected to be inlined constant strings, please replace "(.+)" provided with string`
	key1 := []byte("key1")
	log.Error(err, "message", key1, "value1")                          // want `logging keys are expected to be inlined constant strings, please replace "key1" provided with string`
	log.Error(err, "message", string(key1), "value1")                  // want `logging keys are expected to be inlined constant strings, please replace "string\(key1\)" provided with string`
	log.Error(err, "message", func() bool { return true }(), "value1") // want `logging keys are expected to be inlined constant strings, please replace "(.+)" provided with string`

	type Str string
	key2 := Str("key2")
	log.Error(err, "message", func() string { return "key1" }(), "value1") // want `logging keys are expected to be inlined constant strings, please replace "(.+)" provided with string`
	log.Error(err, "message", "key1", "value1", key2, "value2")            // want `logging keys are expected to be inlined constant strings, please replace "(.+)" provided with string`

	type String = string
	log.Error(err, "message", String(key1), "value1") // want `logging keys are expected to be inlined constant strings, please replace "String\(key1\)" provided with string`

	const Key1Int = 1
	log.Error(err, "message", Key1Int, "value1") // want `logging keys are expected to be inlined constant strings, please replace "Key1Int" provided with string`

	const Key1Str = "key1"
	log.Error(err, "message", Key1Str, "value1")

	log.Error(err, "message", "键1", "value1") // want `logging keys are expected to be alphanumeric strings, please remove any non-latin characters from "键1"`
	const KeyNonASCII = "键1"
	log.Error(err, "message", KeyNonASCII, "value1") // want `logging keys are expected to be alphanumeric strings, please remove any non-latin characters from "键1"`
}
