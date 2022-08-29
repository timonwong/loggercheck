package main

import (
	"golang.org/x/tools/go/analysis"

	"github.com/timonwong/loggercheck"
)

// AnalyzerPlugin provides analyzers as a plugin.
// It follows golangci-lint style plugin.
var AnalyzerPlugin analyzerPlugin

type analyzerPlugin struct{}

func (analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	a, err := loggercheck.NewAnalyzer()
	if err != nil {
		panic(err)
	}
	return []*analysis.Analyzer{
		a,
	}
}
