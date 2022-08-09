package main

import (
	"golang.org/x/tools/go/analysis"

	"github.com/timonwong/logrlint"
)

// AnalyzerPlugin provides analyzers as a plugin.
// It follows golangci-lint style plugin.
var AnalyzerPlugin analyzerPlugin // nolint: deadcode

type analyzerPlugin struct{}

func (analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		logrlint.Analyzer,
	}
}
