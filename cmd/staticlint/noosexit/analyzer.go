package noosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer запрещает прямой вызов os.Exit в функции main пакета main.
var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "forbids direct os.Exit call in main function of package main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Проверяем только пакет main
	if pass.Pkg == nil || pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Name.Name != "main" {
				return true
			}

			// нашли func main()
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				if pkgIdent.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(
						call.Pos(),
						"direct call to os.Exit in main is forbidden; return error instead",
					)
				}

				return true
			})

			return false
		})
	}

	return nil, nil
}
