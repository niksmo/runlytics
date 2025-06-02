// Package osexit provides "os.Exit(n)" direct call detection in main function.
package osexit

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	mainPkg    = "main"
	mainFunc   = "main"
	osPkg      = "os"
	osExitFunc = "Exit"
)

// Analyzer returns instance of [analysis.Analyzer]
func Analyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "osexit",
		Doc:  "Detects direct os.Exit(n) call in main func",
		Run:  run,
	}
}

func run(pass *analysis.Pass) (any, error) {
	mainPkgFiles := getMainPkgFiles(pass.Files, pass.Fset)

	if len(mainPkgFiles) == 0 {
		return nil, nil
	}

	mainFuncDecl := getMainFuncNode(mainPkgFiles)
	if callExpr := findDirectCall(mainFuncDecl); callExpr != nil {
		pass.Reportf(
			callExpr.Pos(),
			"\"os.Exit\" direct call. Use Exit call in nested function or packages.",
		)
	}
	return nil, nil
}

func getMainPkgFiles(passFiles []*ast.File, fset *token.FileSet) []*ast.File {
	lastArg := os.Args[len(os.Args)-1]
	analysingPath, _ := filepath.Abs(lastArg)
	analysingDir := filepath.Dir(analysingPath)

	var files []*ast.File
	for _, file := range passFiles {
		if file.Name.Name == mainPkg {
			filePosition := fset.Position(file.Pos())
			// Exclude tests cache files
			if strings.HasPrefix(filePosition.Filename, analysingDir) {
				files = append(files, file)
			}
		}
	}
	return files
}

func getMainFuncNode(mainPkgFiles []*ast.File) *ast.FuncDecl {
	for _, file := range mainPkgFiles {
		if fd := getMainFuncDecl(file.Decls); fd != nil {
			return fd
		}
	}
	return nil
}

func getMainFuncDecl(decls []ast.Decl) *ast.FuncDecl {
	for _, decl := range decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fd.Name.Name == mainFunc {
			return fd
		}
	}
	return nil
}

func findDirectCall(funcDecl *ast.FuncDecl) *ast.CallExpr {
	var ret *ast.CallExpr
	ast.Inspect(funcDecl, func(n ast.Node) bool {
		switch e := n.(type) {
		case *ast.FuncDecl, *ast.BlockStmt, *ast.ExprStmt:
			return true
		case *ast.CallExpr:
			if isOSExitCall(e) {
				ret = e
			}
			return false
		default:
			return false
		}
	})

	return ret
}

func isOSExitCall(e *ast.CallExpr) bool {
	se, ok := e.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	i, ok := se.X.(*ast.Ident)
	if !ok {
		return false
	}
	if i.Name == osPkg && se.Sel.Name == osExitFunc {
		return true
	}
	return false
}
