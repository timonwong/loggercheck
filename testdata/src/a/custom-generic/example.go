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

func ExampleGenericCustomOnly() {
	l := zap.NewExample().Sugar()
	logger := &Logger[int8, int]{l}

	logger.Infow("custom message", "hello")                                     // want `odd number of arguments passed as key-value pairs for logging`
	logger.Debugw("embedded sugar log also works without custom rules", "key1") // want `odd number of arguments passed as key-value pairs for logging`
}
