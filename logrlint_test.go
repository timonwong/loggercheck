package logrlint_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/timonwong/logrlint"
)

func TestAll(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, logrlint.NewAnalyzer(), "a/all")
}

func TestKlogOnly(t *testing.T) {
	testdata := analysistest.TestData()
	a := logrlint.NewAnalyzer()

	testCases := []struct {
		name  string
		flags []string
	}{
		{
			name:  "disable-all-then-enable-klog",
			flags: []string{"-disableall", "-enable=klog"},
		},
		{
			name:  "just-disable-logr",
			flags: []string{"-disable=klog"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := a.Flags.Parse(tc.flags)
			require.NoError(t, err)
			analysistest.Run(t, testdata, a, "a/klogonly")
		})
	}
}
