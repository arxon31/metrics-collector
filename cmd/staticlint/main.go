package main

import (
	mylinter "github.com/arxon31/metrics_linter"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"strings"
)

func main() {

	var myCheckers []*analysis.Analyzer

	// Some analyzers from analysis/passes
	myCheckers = append(myCheckers,
		shadow.Analyzer,
		printf.Analyzer,
		structtag.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer)

	// All SA analyzers from staticcheck.io
	for _, v := range staticcheck.Analyzers {
		if strings.Contains(v.Analyzer.Name, "SA") {
			myCheckers = append(myCheckers, v.Analyzer)
		}
	}

	// One ST analyzer from staticcheck.io
	for _, v := range stylecheck.Analyzers {
		myCheckers = append(myCheckers, v.Analyzer)
		break
	}

	myCheckers = append(myCheckers, mylinter.OsExitChecker)

	multichecker.Main(
		myCheckers...,
	)
}
