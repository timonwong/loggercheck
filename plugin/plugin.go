package main

import (
	"github.com/timonwong/logrlint/pkg/analyzer"
	"golang.org/x/tools/go/analysis"
)

type analyzerPlugin struct{}

func (*analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		analyzer.Analyzer,
	}
}

//goland:noinspection GoUnusedGlobalVariable
var AnalyzerPlugin analyzerPlugin
