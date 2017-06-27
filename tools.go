package atool

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type Struct struct {
	Name       string
	Definition *ast.StructType
}

func StructMapFile(filename string) ([]Struct, error) {
	tokens := token.NewFileSet()
	file, err := parser.ParseFile(tokens, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	var res []Struct
	for _, node := range file.Decls {
		res = append(res, StructMap(node)...)
	}
	return res, nil
}

func StructMap(decls ...ast.Node) []Struct {
	var res []Struct
	var stack []ast.Node
	var name string
	for i := len(decls) - 1; i >= 0; i-- {
		stack = append(stack, decls[i])
	}

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		switch v := node.(type) {
		case *ast.TypeSpec:
			name = v.Name.Name
			stack = append(stack, v.Type)
		case *ast.StructType:
			res = append(res, Struct{name, v})
		case *ast.GenDecl:
			for _, spec := range v.Specs {
				stack = append(stack, spec)
			}
		}
	}
	return res
}
