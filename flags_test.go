package loggercheck

import (
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerCheckersFlag(t *testing.T) {
	testCases := []struct {
		name      string
		flagValue string
		wantError string
		want      []string
	}{
		{
			name:      "empty",
			flagValue: "",
			want:      nil,
		},
		{
			name:      "klog",
			flagValue: "klog",
			want:      []string{"klog"},
		},
		{
			name:      "klog-and-logr",
			flagValue: "logr,klog",
			want:      []string{"klog", "logr"},
		},
		{
			name:      "invalid-logger",
			flagValue: "klog,logr,xxx",
			wantError: "-ignoredloggers: unknown logger: \"xxx\"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := loggerCheckersFlag{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			fs.Var(&f, "ignoredloggers", "")

			err := fs.Parse([]string{"-ignoredloggers=" + tc.flagValue})
			if tc.wantError != "" {
				assert.ErrorContains(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, f.List())
			}
		})
	}
}

func TestRuleFileFlag_NoRuleFile(t *testing.T) {
	f := ruleFileFlag{}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&f, "rulefile", "")

	err := fs.Parse([]string{"-rulefile=testdata/xxx-not-exists-xxx.txt"})
	assert.ErrorContains(t, err, "open testdata/xxx-not-exists-xxx.txt: no such file or directory")
}

func TestRuleFileFlag_WrongRuleFile(t *testing.T) {
	f := ruleFileFlag{}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&f, "rulefile", "")

	err := fs.Parse([]string{"-rulefile=testdata/wrong-rules.txt"})
	assert.ErrorContains(t, err, "error parse rule at line 2: invalid rule format")
}
