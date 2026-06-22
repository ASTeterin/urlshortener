package main

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var Analyzer = &analysis.Analyzer{
	Name: "panicexit",
	Doc:  "reports panic usage and log.Fatal/os.Exit calls outside main()",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Body == nil {
				continue
			}

			isMain := funcDecl.Name.Name == "main"

			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
					pass.Reportf(call.Pos(), "use of built-in panic")
				}

				if !isMain {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						if pkg, ok := sel.X.(*ast.Ident); ok {
							if (pkg.Name == "log" && sel.Sel.Name == "Fatal") ||
								(pkg.Name == "os" && sel.Sel.Name == "Exit") {
								pass.Reportf(call.Pos(), "use of %s.%s outside main function", pkg.Name, sel.Sel.Name)
							}
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}

func main() {
	singlechecker.Main(Analyzer)
}
