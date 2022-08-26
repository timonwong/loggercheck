package loggercheck_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/timonwong/loggercheck"
)

func TestLinter(t *testing.T) {
	testdata := analysistest.TestData()

	testCases := []struct {
		name     string
		patterns string
		flags    []string
	}{
		{
			name:     "all",
			patterns: "a/all",
		},
		{
			name:     "klogonly",
			patterns: "a/klogonly",
			flags:    []string{"-disable=logr,zap"},
		},
		{
			name:     "custom-only",
			patterns: "a/customonly",
			flags: []string{
				"-disable=klog,logr,zap",
				"-logger=mylogger:a/customonly:" +
					"(*a/customonly.Logger).Debugw," +
					"(*a/customonly.Logger).Infow," +
					"(*a/customonly.Logger).Warnw," +
					"(*a/customonly.Logger).Errorw," +
					"(*a/customonly.Logger).With," +
					"a/customonly.Debugw," +
					"a/customonly.Infow," +
					"a/customonly.Warnw," +
					"a/customonly.Errorw," +
					"a/customonly.With",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		a := loggercheck.NewAnalyzer()
		t.Run(tc.name, func(t *testing.T) {
			err := a.Flags.Parse(tc.flags)
			require.NoError(t, err)
			analysistest.Run(t, testdata, a, tc.patterns)
		})
	}
}

func TestOptions(t *testing.T) {
	testdata := analysistest.TestData()

	customLogger := loggercheck.WithCustomLogger("mylogger", "a/customonly",
		[]string{
			"(*a/customonly.Logger).Debugw",
			"(*a/customonly.Logger).Infow",
			"(*a/customonly.Logger).Warnw",
			"(*a/customonly.Logger).Errorw",
			"(*a/customonly.Logger).With",

			"a/customonly.Debugw",
			"a/customonly.Infow",
			"a/customonly.Warnw",
			"a/customonly.Errorw",
			"a/customonly.With",
		})

	testCases := []struct {
		name    string
		options []loggercheck.Option
	}{
		{
			name: "disable-all-then-enable-mylogger",
			options: []loggercheck.Option{
				loggercheck.WithDisableFlags(true),
				customLogger,
				loggercheck.WithDisable([]string{"klog", "logr", "zap"}),
			},
		},
		{
			name: "ignore-logr",
			options: []loggercheck.Option{
				loggercheck.WithDisableFlags(true),
				customLogger,
				loggercheck.WithDisable([]string{"logr"}),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			a := loggercheck.NewAnalyzer(tc.options...)
			analysistest.Run(t, testdata, a, "a/customonly")
		})
	}
}
