package logrlint_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/timonwong/logrlint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, logrlint.NewAnalyzer(), "a/all")
}

func TestKlogOnly(t *testing.T) {
	testdata := analysistest.TestData()
	a := logrlint.NewAnalyzer()
	err := a.Flags.Parse([]string{"-ignoredloggers=logr"})
	require.NoError(t, err)
	analysistest.Run(t, testdata, a, "a/klogonly")
}
