package atool

import (
	"go/ast"
	"go/parser"
	"go/token"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type Struct struct {
	Name       string
	Definition *ast.StructType `json:"-"`
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
	Name    string
	Type    ast.Expr
	printer *Printer
}

func (u *Arg) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name       string
		GolangType string
		IsError    bool
	}{

		Name:       u.Name,
		GolangType: u.GolangType(),
		IsError:    u.IsError(),
	})
}

func (arg *Arg) IsError() bool {
	v, ok := arg.Type.(*ast.Ident)
	return ok && v.Name == "error"
}

func (arg *Arg) GolangType() string {
	return arg.printer.ToString(arg.Type)
}

type Method struct {
	Name string
	In   []Arg `json:",omitempty"`
	Out  []Arg `json:",omitempty"`
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
	Methods    []Method           `json:",omitempty"`
	Definition *ast.InterfaceType `json:"-"`
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
	Imports    map[string]string `json:",omitempty"`
	Interfaces []Interface       `json:",omitempty"`
	Structs    []Struct          `json:",omitempty"`
	Printer    *Printer          `json:"-"`
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

	printer := &Printer{string(content)}

	var structs []Struct
	for _, node := range file.Decls {
		structs = append(structs, Structs(node)...)
	}
	var interfaces []Interface
	for _, node := range file.Decls {
		interfaces = append(interfaces, Interfaces(printer, node)...)
	}

	imports := make(map[string]string)
	for _, imp := range file.Imports {
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		imports[imp.Path.Value] = alias
	}
	return &File{
		Package:    file.Name.Name,
		Printer:    printer,
		Structs:    structs,
		Interfaces: interfaces,
		Imports:    imports,
	}, nil

}

func InterfacesFile(filename string) ([]Interface, *Printer, error) {
	tokens := token.NewFileSet()
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	printer := &Printer{string(content)}
	file, err := parser.ParseFile(tokens, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, nil, err
	}
	var res []Interface
	for _, node := range file.Decls {
		res = append(res, Interfaces(printer, node)...)
	}
	return res, &Printer{string(content)}, nil
}

func Interfaces(printer *Printer, decls ...ast.Node) []Interface {
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
				iface.Methods = append(iface.Methods, AsMethod(m, printer))
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

func AsMethod(m *ast.Field, printer *Printer) Method {
	method := Method{Name: m.Names[0].Name}
	def := m.Type.(*ast.FuncType)
	if def.Params != nil {
		for i, p := range def.Params.List {
			if p.Names != nil {
				for _, name := range p.Names {
					method.In = append(method.In, Arg{name.Name, p.Type, printer})
				}
			} else {
				method.In = append(method.In, Arg{fmt.Sprintf("arg%v", i), p.Type, printer})
			}
		}
	}
	if def.Results != nil {
		for i, p := range def.Results.List {
			name := fmt.Sprintf("ret%v", i)
			if p.Names != nil {
				name = p.Names[0].Name
			}
			method.Out = append(method.Out, Arg{name, p.Type, printer})
		}
	}
	return method
}
