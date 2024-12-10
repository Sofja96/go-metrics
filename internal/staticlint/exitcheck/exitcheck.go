package exitcheck

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// ExitAnalyzer анализирует вызовы os.Exit в функции main пакета main.
var ExitAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check os.Exit in main func of main pkg",
	Run:  run,
}

// run выполняет проверку для каждого файла.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Проверяем, является ли текущий файл файлом пакета main
		if pass.Pkg.Name() != "main" {
			continue
		}

		// Ищем функцию main
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != "main" || funcDecl.Recv != nil {
				continue
			}

			// Проверяем тело функции main на наличие вызовов os.Exit
			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				callExpr, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Проверяем, что вызывается os.Exit
				if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "os" && selExpr.Sel.Name == "Exit" {
						pass.Reportf(callExpr.Pos(), "os.Exit call in main")
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
