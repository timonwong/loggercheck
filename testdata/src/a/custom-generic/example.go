package custom_generic

import "go.uber.org/zap"

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Integer interface {
	Signed | Unsigned
}
type Logger[T, W Integer] struct {
	*zap.SugaredLogger
}

func (l *Logger[T, W]) Infow(message string, keysAndValues ...any) {
	l.SugaredLogger.Infow(message, keysAndValues...)
}

func (l *Logger[T, W]) Infox(message string, keysAndValues ...string) {
	kvs := make([]any, 0, len(keysAndValues))
	for _, v := range keysAndValues {
		kvs = append(kvs, v)
	}
	l.SugaredLogger.Infow(message, kvs...)
}

func ExampleGenericCustomOnly() {
	l := zap.NewExample().Sugar()
	logger := &Logger[int8, int]{l}

	logger.Infox("will not check this", "hello", "world")
	logger.Infow("custom message", "hello")                                     // want `odd number of arguments passed as key-value pairs for logging`
	logger.Debugw("embedded sugar log also works without custom rules", "key1") // want `odd number of arguments passed as key-value pairs for logging`
}
