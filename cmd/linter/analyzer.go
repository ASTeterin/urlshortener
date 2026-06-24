package main

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "linter",
	Doc:  "Checks for panic usage and log.Fatal/os.Exit outside main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		pkgName := file.Name.Name
		var inMainFunc bool

		var walk func(n ast.Node)
		walk = func(n ast.Node) {
			if n == nil {
				return
			}
			switch node := n.(type) {
			case *ast.FuncDecl:
				prev := inMainFunc
				inMainFunc = pkgName == "main" && node.Name.Name == "main"
				if node.Body != nil {
					walk(node.Body)
				}
				inMainFunc = prev
			case *ast.CallExpr:
				if ident, ok := node.Fun.(*ast.Ident); ok && ident.Name == "panic" {
					pass.Reportf(node.Pos(), "use of built-in panic")
				}
				if !inMainFunc {
					if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
						if xIdent, ok := sel.X.(*ast.Ident); ok {
							if (xIdent.Name == "log" && sel.Sel.Name == "Fatal") ||
								(xIdent.Name == "os" && sel.Sel.Name == "Exit") {
								pass.Reportf(node.Pos(), "call to %s.%s outside main.main", xIdent.Name, sel.Sel.Name)
							}
						}
					}
				}
			default:
				// Для всех остальных узлов рекурсивно обходим детей
				// Используем ast.Inspect, но отключаем его рекурсию, чтобы управлять контекстом вручную
				ast.Inspect(n, func(child ast.Node) bool {
					if child == n {
						return true
					}
					walk(child)
					return false // Прерываем рекурсию ast.Inspect, так как мы сами вызвали walk
				})
			}
		}
		walk(file)
	}
	return nil, nil
}
