package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/timonwong/loggercheck"
)

func main() {
	a, err := loggercheck.NewAnalyzer()
	if err != nil {
		panic(err)
	}
	singlechecker.Main(a)
}
