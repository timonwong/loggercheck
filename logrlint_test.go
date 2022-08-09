package logrlint_test

import (
	"testing"

	"github.com/timonwong/logrlint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, logrlint.Analyzer, "a")
}
