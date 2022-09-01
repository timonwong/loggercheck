package custom_generic

import "go.uber.org/zap"

type Logger[T any] struct {
	*zap.SugaredLogger
}

func (l *Logger[T]) Infow(message string, keysAndValues ...any) {
	l.SugaredLogger.Infow(message, keysAndValues...)
}

func ExampleGenericCustomOnly() {
	l := zap.NewExample().Sugar()
	logger := &Logger[any]{l}

	logger.Infow("custom message", "hello")                                     // want `odd number of arguments passed as key-value pairs for logging`
	logger.Debugw("embedded sugar log also works without custom rules", "key1") // want `odd number of arguments passed as key-value pairs for logging`
}
