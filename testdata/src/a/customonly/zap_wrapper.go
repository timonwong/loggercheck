package customonly

import "go.uber.org/zap"

var l = New()

type Logger struct {
	s *zap.SugaredLogger
}

func New() *Logger {
	logger := zap.NewExample().Sugar()
	return &Logger{s: logger}
}

func (l *Logger) With(keysAndValues ...interface{}) *Logger {
	return &Logger{
		s: l.s.With(keysAndValues...),
	}
}

func (l *Logger) XXXDebugw(msg string, keysAndValues ...interface{}) {
	l.s.Debugw(msg, keysAndValues...)
}

func (l *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	l.s.Debugw(msg, keysAndValues...)
}

func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.s.Infow(msg, keysAndValues...)
}

func (l *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	l.s.Warnw(msg, keysAndValues...)
}

func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.s.Errorw(msg, keysAndValues...)
}

func (l *Logger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.s.Fatalw(msg, keysAndValues...)
}

func (l *Logger) Sync() error {
	return l.s.Sync()
}

// package level wrap func

func With(keysAndValues ...interface{}) *Logger {
	return &Logger{
		s: l.s.With(keysAndValues...),
	}
}

func Debugw(msg string, keysAndValues ...interface{}) {
	l.s.Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	l.s.Infow(msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	l.s.Warnw(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	l.s.Errorw(msg, keysAndValues...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	l.s.Fatalw(msg, keysAndValues...)
}

func Sync() error {
	return l.s.Sync()
}
