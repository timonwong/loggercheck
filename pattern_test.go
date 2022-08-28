package loggercheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePattern(t *testing.T) {
	testCases := []struct {
		name      string
		pattern   string
		wantError error
		want      Pattern
	}{
		{
			name:      "invalid-pattern-missing-paren",
			pattern:   "(*go.uber.org/zap/SugaredLogger.Debugw",
			wantError: errInvalidPattern,
		},
		{
			name:      "invalid-pattern-receiver-no-type",
			pattern:   "(*go.uber.org/zap/SugaredLogger).Debugw",
			wantError: errInvalidPattern,
		},
		{
			name:      "invalid-pattern-just-import",
			pattern:   "go.uber.org/zap",
			wantError: errInvalidPattern,
		},
		{
			name:    "zap",
			pattern: "(*go.uber.org/zap.SugaredLogger).Debugw",
			want: Pattern{
				IsReceiver:    true,
				PackageImport: "go.uber.org/zap",
				ReceiverType:  "*SugaredLogger",
				FuncName:      "Debugw",
			},
		},
		{
			name:    "klog-no-receiver",
			pattern: "k8s.io/klog/v2.InfoS",
			want: Pattern{
				PackageImport: "k8s.io/klog/v2",
				FuncName:      "InfoS",
			},
		},
		{
			name:    "logr",
			pattern: "(github.com/go-logr/logr.Logger).Error",
			want: Pattern{
				IsReceiver:    true,
				PackageImport: "github.com/go-logr/logr",
				ReceiverType:  "Logger",
				FuncName:      "Error",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parsePattern(tc.pattern)
			if tc.wantError != nil {
				assert.EqualError(t, err, tc.wantError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
