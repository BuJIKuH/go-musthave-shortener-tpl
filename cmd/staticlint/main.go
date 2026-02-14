package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"

	// стандартные passes
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	// staticcheck
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	// публичные анализаторы
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"

	// наш анализатор
	"github.com/BuJIKuH/go-musthave-shortener-tpl/cmd/staticlint/noosexit"
)

func main() {
	var analyzers = []*analysis.Analyzer{
		// стандартные
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,

		// публичные
		errcheck.Analyzer,
		ineffassign.Analyzer,

		// кастомный
		noosexit.Analyzer,
	}

	// SA — обязательно все
	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	// не SA (требование задания)
	for _, a := range stylecheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}
	for _, a := range quickfix.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	multichecker.Main(analyzers...)
}
