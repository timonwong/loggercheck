package loggercheck_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/timonwong/loggercheck"
)

func TestLinter(t *testing.T) {
	testdata := analysistest.TestData()

	testConfig := []byte(`# loggercheck sample config
disable:
    - klog
    - logr
    - zap
custom-checkers:
    - name: mylogger
      package-import: a/customonly
      funcs:
        - (*a/customonly.Logger).Debugw
        - (*a/customonly.Logger).Infow
        - (*a/customonly.Logger).Warnw
        - (*a/customonly.Logger).Errorw
        - (*a/customonly.Logger).With
        - a/customonly.Debugw
        - a/customonly.Infow
        - a/customonly.Warnw
        - a/customonly.Errorw
        - a/customonly.With
`)

	f, err := os.CreateTemp("/tmp", "loggercheck-test-cfg-")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(testConfig); err != nil {
		t.Errorf("write test config file failed: %v", err)
	}

	testConfigFile := f.Name()
	t.Cleanup(func() {
		f.Close()
		os.Remove(testConfigFile)
	})

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
				"-config=" + testConfigFile,
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

	customLogger := loggercheck.WithConfig(&loggercheck.Config{
		Disable: []string{"klog", "logr", "zap"},
		CustomCheckers: []loggercheck.Checker{
			{
				Name:          "mylogger",
				PackageImport: "a/customonly",
				Funcs: []string{
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
				},
			},
		},
	})

	testCases := []struct {
		name    string
		options []loggercheck.Option
	}{
		{
			name: "disable-all-then-enable-mylogger",
			options: []loggercheck.Option{
				customLogger,
			},
		},
		{
			name: "ignore-logr",
			options: []loggercheck.Option{
				customLogger,
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
