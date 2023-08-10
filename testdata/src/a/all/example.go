package all

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	kitlog "github.com/go-kit/log"
	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"k8s.io/klog/v2"
)

func ExampleInvalid() {
	// function pointer is not supported

	log := logr.Discard()
	logFn := log.Info
	logFn("message", "key1") // cannot be detected
}

func ExampleLogr() {
	err := fmt.Errorf("error")

	log := logr.Discard()
	log = log.WithValues("key")                                     // want `odd number of arguments passed as key-value pairs for logging`
	log.Info("message", "key1", "value1", "key2", "value2", "key3") // want `odd number of arguments passed as key-value pairs for logging`
	log.Error(err, "message", "key1", "value1", "key2")             // want `odd number of arguments passed as key-value pairs for logging`
	log.Error(err, "message", "key1", "value1", "key2", "value2")

	var log2 logr.Logger
	log2 = log
	log2.Info("message", "key1") // want `odd number of arguments passed as key-value pairs for logging`

	log3 := logr.FromContextOrDiscard(context.TODO())
	log3.Error(err, "message", "key1") // want `odd number of arguments passed as key-value pairs for logging`
	args := []interface{}{"abc"}
	log3.Error(err, "message", args...) // not supported
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
	logger.Info("message", "key1") // want `odd number of arguments passed as key-value pairs for logging`
}

func ExampleZap() {
	err := errors.New("example error")

	field := zap.String("key1", "value1")
	field2 := zap.Int("key2", 2)
	field3 := zap.Bool("key3", true)

	// custom SugaredLogger
	log := zap.NewExample().Sugar()
	defer log.Sync()

	log.Infow("abc", "key1", "value1")
	log.Infow("abc", "key1", "value1", "key2") // want `odd number of arguments passed as key-value pairs for logging`

	log.Errorw("message", "err", err, "key1", "value1")
	log.Errorw("message", err, "key1", "value1", "key2", "value2") // want `odd number of arguments passed as key-value pairs for logging`

	// with test
	log.With("with_key1", "with_value1").Infow("message", "key1", "value1")
	log.With("with_key1", "with_value1", field).Infow("message", "key1", "value1")
	log.With("with_key1", "with_value1", field, field2, field3).Infow("message", "key1", "value1")
	log.With("with_key1", "with_value1").Infow("message", "key1", "value1", "key2") // want `odd number of arguments passed as key-value pairs for logging`
	log.With("with_key1").Infow("message", "key1", "value1")                        // want `odd number of arguments passed as key-value pairs for logging`
	log.With("with_key1", field).Infow("message", "key1", "value1")                 // want `odd number of arguments passed as key-value pairs for logging`

	// default global SugaredLogger
	zap.S().With("with_key1", "with_value1").Infow("message", "key1", "value1", "key2", "value2")
	zap.S().With("with_key1", "with_value1", field).Infow("message")
	zap.S().With("with_key1", field).Infow("message")                    // want `odd number of arguments passed as key-value pairs for logging`
	zap.S().Infow("message", "key1", "value1", "key2", "value2", "key3") // want `odd number of arguments passed as key-value pairs for logging`

	zap.S().Infow("message", zap.String("key1", "value1"))
	zap.S().Infow("message", field, field2, field3)
	zap.S().Infow("message", field, "key1", "value1")
	zap.S().Infow("message", field, field2, field3, "key1") // want `odd number of arguments passed as key-value pairs for logging`

	zap.S().Errorw("message", "err", err, "key1", "value1")
	zap.S().Errorw("message", err, "message", "key1") // want `odd number of arguments passed as key-value pairs for logging`
}

func ExampleGokitLog() {
	logger := kitlog.NewNopLogger()

	logger.Log("msg", "message", "key1", "value1")
	logger.Log("msg")                    // want `odd number of arguments passed as key-value pairs for logging`
	logger.Log("msg", "message", "key1") // want `odd number of arguments passed as key-value pairs for logging`

	kitlog.With(logger, "key1", "value1").Log("msg", "message")
	kitlog.With(logger, "key1").Log("msg", "message") // want `odd number of arguments passed as key-value pairs for logging`
}

func ExampleSlog() {
	logger := slog.With("key1", "value1")

	slog.Info("msg", "key1", "value1")
	logger.Info("msg", "key1", "value1")
	logger.Info("msg", "key1")                        // want `odd number of arguments passed as key-value pairs for logging`
	slog.Info("msg", "key1")                          // want `odd number of arguments passed as key-value pairs for logging`
	logger.InfoContext(context.TODO(), "msg", "key1") // want `odd number of arguments passed as key-value pairs for logging`

	slog.With("key1", "value1").Info("msg")
	slog.With("key1", "value1").Info("msg", "key1") // want `odd number of arguments passed as key-value pairs for logging`

	slog.Info("with group ok", "key1", "value1", slog.Group("group_val_key", "gkey1", "gvalue1", "gkey2", "gvalue2"))
	slog.Info("with group missing val", "key1", "value1", slog.Group("group_val_key", "gkey1", "gvalue1", "gkey2")) // want `odd number of arguments passed as key-value pairs for logging`
}
