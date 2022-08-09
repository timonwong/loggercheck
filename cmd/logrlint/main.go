package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/timonwong/logrlint"
)

func main() {
	singlechecker.Main(logrlint.Analyzer)
}
