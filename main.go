package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
)

type Array struct {
	Type ast.Expr
}

func main() {
	tokens := token.NewFileSet()
	file, err := parser.ParseFile(tokens, "sample.go", nil, parser.AllErrors)
	if err != nil {
		panic(err)
	}
	for name, def := range findStructs(file.Decls) {
		fmt.Println("Struct", name)
		for _, field := range def.Fields.List {
			fmt.Println("  ", field.Names[0].Name)

			if f, ok := field.Type.(*ast.ArrayType); ok {
				fmt.Println("    item:", f.Elt)
			} else {
				fmt.Println("    ", field.Type)
			}
		}
	}
}

func findStructs(decls []ast.Decl) map[string]*ast.StructType {
	var res map[string]*ast.StructType = make(map[string]*ast.StructType)
	var queue []ast.Node
	var name string
	for _, decl := range decls {
		queue = append(queue, decl)
	}

	for len(queue) > 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		switch v := node.(type) {
		case *ast.TypeSpec:
			name = v.Name.Name
			queue = append(queue, v.Type)
		case *ast.StructType:
			res[name] = v
		case *ast.GenDecl:
			for _, spec := range v.Specs {
				queue = append(queue, spec)
			}
		default:
			fmt.Println("?", reflect.TypeOf(v).Elem().Name())
		}
	}
	return res
}
