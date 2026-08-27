package issue108

import (
	"context"
	"log/slog"
)

func ExampleMultiValuedHelper() {
	slog.ErrorContext(fn())
	slog.InfoContext(fn())
	slog.Default().WarnContext(fn())
	slog.Default().ErrorContext(fn())
	slog.Info(keyValues())
	slog.Default().Info(keyValues())
}

func fn() (context.Context, string) {
	return nil, "test"
}

func keyValues() (string, string, string) {
	return "msg", "key", "value"
}
