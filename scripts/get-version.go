// +build ignore

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

func main() {
	// version.go ファイルをパース
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "version.go", nil, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing version.go: %v\n", err)
		os.Exit(1)
	}

	// Version 定数を探す
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, name := range valueSpec.Names {
				if name.Name == "Version" {
					if i < len(valueSpec.Values) {
						if lit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
							if lit.Kind == token.STRING {
								// ダブルクォートを除去
								version := lit.Value[1 : len(lit.Value)-1]
								fmt.Print(version)
								os.Exit(0)
							}
						}
					}
				}
			}
		}
	}

	fmt.Fprintf(os.Stderr, "Version constant not found\n")
	os.Exit(1)
}