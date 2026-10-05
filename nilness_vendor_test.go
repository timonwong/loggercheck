package loggercheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/timonwong/loggercheck"
)

func TestStringerNilnessVendoredNames(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), loggercheck.NewAnalyzer(), "x")
}
