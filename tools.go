package atool

import (
	"go/ast"
	"go/parser"
	"go/token"
	"fmt"
	"io/ioutil"
)

type Struct struct {
	Name       string
	Definition *ast.StructType
}

func StructsFile(filename string) ([]Struct, *Printer, error) {
	tokens := token.NewFileSet()
	file, err := parser.ParseFile(tokens, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, nil, err
	}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	var res []Struct
	for _, node := range file.Decls {
		res = append(res, Structs(node)...)
	}
	return res, &Printer{string(content)}, nil
}

func Structs(decls ...ast.Node) []Struct {
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

type Arg struct {
	Name string
	Type ast.Expr
}

func (arg *Arg) IsError() bool {
	return arg.Type.(*ast.Ident).Name == "error"
}

type Method struct {
	Name string
	In   []Arg
	Out  []Arg
}

func (m *Method) HasInput() bool {
	return len(m.In) > 0
}

func (m *Method) HasOutput() bool {
	return len(m.Out) > 0
}

func (m *Method) ErrorOutputs() []Arg {
	var args []Arg
	for _, arg := range m.Out {
		if arg.IsError() {
			args = append(args, arg)
		}
	}
	return args
}

func (m *Method) NonErrorOutputs() []Arg {
	var args []Arg
	for _, arg := range m.Out {
		if !arg.IsError() {
			args = append(args, arg)
		}
	}
	return args
}

type Interface struct {
	Name       string
	Methods    []Method
	Definition *ast.InterfaceType
}

type Printer struct {
	Src string
}

func (p *Printer) ToString(node ast.Node) string {
	return p.Src[node.Pos()-1:node.End()-1]
}

func (in *Interface) Method(name string) *Method {
	for _, m := range in.Methods {
		if m.Name == name {
			return &m
		}
	}
	return nil
}

type File struct {
	Package    string
	Interfaces []Interface
	Structs    []Struct
	Printer    *Printer
}

func Scan(filename string) (*File, error) {
	tokens := token.NewFileSet()
	file, err := parser.ParseFile(tokens, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var structs []Struct
	for _, node := range file.Decls {
		structs = append(structs, Structs(node)...)
	}
	var interfaces []Interface
	for _, node := range file.Decls {
		interfaces = append(interfaces, Interfaces(node)...)
	}

	return &File{
		Package:    file.Name.Name,
		Printer:    &Printer{string(content)},
		Structs:    structs,
		Interfaces: interfaces,
	}, nil

}

func InterfacesFile(filename string) ([]Interface, *Printer, error) {
	tokens := token.NewFileSet()
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	file, err := parser.ParseFile(tokens, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, nil, err
	}
	var res []Interface
	for _, node := range file.Decls {
		res = append(res, Interfaces(node)...)
	}
	return res, &Printer{string(content)}, nil
}

func Interfaces(decls ...ast.Node) []Interface {
	var res []Interface
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
		case *ast.InterfaceType:
			iface := Interface{Name: name, Definition: v}
			for _, m := range v.Methods.List {
				iface.Methods = append(iface.Methods, AsMethod(m))
			}
			res = append(res, iface)
		case *ast.GenDecl:
			for _, spec := range v.Specs {
				stack = append(stack, spec)
			}
		}
	}
	return res
}

func AsMethod(m *ast.Field) Method {
	method := Method{Name: m.Names[0].Name}
	def := m.Type.(*ast.FuncType)
	if def.Params != nil {
		for i, p := range def.Params.List {
			name := fmt.Sprintf("arg%v", i)
			if p.Names != nil {
				name = p.Names[0].Name
			}
			method.In = append(method.In, Arg{name, p.Type})
		}
	}
	if def.Results != nil {
		for i, p := range def.Results.List {
			name := fmt.Sprintf("ret%v", i)
			if p.Names != nil {
				name = p.Names[0].Name
			}
			method.Out = append(method.Out, Arg{name, p.Type})
		}
	}
	return method
}
