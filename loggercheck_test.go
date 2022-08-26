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
	}

	a := loggercheck.NewAnalyzer()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := a.Flags.Parse(tc.flags)
			require.NoError(t, err)
			analysistest.Run(t, testdata, a, tc.patterns)
		})
	}
}
