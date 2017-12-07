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
	Comment    string          `json:",omitempty"`
	Fields     []*Arg
	Definition *ast.StructType `json:"-"`
	printer    *Printer        `json:"-"`
}

func StructsFile(filename string) ([]*Struct, *Printer, error) {
	tokens := token.NewFileSet()
	file, err := parser.ParseFile(tokens, filename, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	printer := &Printer{string(content), ast.NewCommentMap(tokens, file, file.Comments)}
	var res []*Struct
	for _, node := range file.Decls {
		res = append(res, Structs(printer, node)...)
	}
	return res, printer, nil
}

func Structs(printer *Printer, decls ...ast.Node) []*Struct {
	var res []*Struct
	var stack []ast.Node
	var name string
	for i := len(decls) - 1; i >= 0; i-- {
		stack = append(stack, decls[i])
	}

	var lastComment string
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		switch v := node.(type) {
		case *ast.TypeSpec:
			name = v.Name.Name
			stack = append(stack, v.Type)
		case *ast.StructType:
			res = append(res, &Struct{name, lastComment, getArgs(printer, v.Fields.List), v, printer,})
		case *ast.GenDecl:
			lastComment = joinComments(printer.CommentMap[v])
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
	Comment string
	printer *Printer
}

func (u *Arg) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name       string
		GolangType string
		Comment    string `json:",omitempty"`
		IsError    bool
	}{

		Name:       u.Name,
		GolangType: u.GolangType(),
		Comment:    u.Comment,
		IsError:    u.IsError(),
	})
}

func (arg *Arg) IsSimple() bool {
	v, ok := arg.Type.(*ast.Ident)
	if ok {
		switch(v.Name) {
		case "byte", "rune", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint64", "string", "bool", "float32", "float64":
			return true
		}
	}
	return false
}

func (arg *Arg) IsInteger() bool {
	v, ok := arg.Type.(*ast.Ident)
	if ok {
		switch(v.Name) {
		case "byte", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint64":
			return true
		}
	}
	return false
}

func (arg *Arg) IsFloat() bool {
	v, ok := arg.Type.(*ast.Ident)
	if ok {
		switch(v.Name) {
		case "float32", "float64":
			return true
		}
	}
	return false
}

func (arg *Arg) IsString() bool {
	v, ok := arg.Type.(*ast.Ident)
	if ok {
		switch(v.Name) {
		case "string":
			return true
		}
	}
	return false
}

func (arg *Arg) IsBoolean() bool {
	v, ok := arg.Type.(*ast.Ident)
	if ok {
		switch(v.Name) {
		case "bool":
			return true
		}
	}
	return false
}
func (arg *Arg) IsArray() bool {
	_, ok := arg.Type.(*ast.ArrayType)
	return ok
}
func (arg *Arg) ArrayItem() *Arg {
	v, ok := arg.Type.(*ast.ArrayType)
	if !ok {
		return nil
	}
	return &Arg{
		Name:    "",
		Type:    v.Elt,
		Comment: "",
		printer: arg.printer,
	}
}
func (arg *Arg) IsMap() bool {
	_, ok := arg.Type.(*ast.MapType)
	return ok
}
func (arg *Arg) IsError() bool {
	v, ok := arg.Type.(*ast.Ident)
	return ok && v.Name == "error"
}

func (arg *Arg) GolangType() string {
	return arg.printer.ToString(arg.Type)
}

type Value struct {
	Name    string
	Type    ast.Expr
	Comment string
	printer *Printer
	Value   ast.Expr
}

func (arg *Value) GolangValue() string {
	return arg.printer.ToString(arg.Value)
}

func (u *Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name        string
		GolangType  string
		Comment     string `json:",omitempty"`
		GolangValue string
		IsError     bool
	}{

		Name:        u.Name,
		GolangType:  u.GolangType(),
		GolangValue: u.GolangValue(),
		Comment:     u.Comment,
		IsError:     u.IsError(),
	})
}

func (arg *Value) IsError() bool {
	v, ok := arg.Type.(*ast.Ident)
	return ok && v.Name == "error"
}

func (arg *Value) GolangType() string {
	return arg.printer.ToString(arg.Type)
}

type Method struct {
	Name    string
	Comment string `json:",omitempty"`
	In      []*Arg  `json:",omitempty"`
	Out     []*Arg  `json:",omitempty"`
}

func (m *Method) HasInput() bool {
	return len(m.In) > 0
}

func (m *Method) HasOutput() bool {
	return len(m.Out) > 0
}

func (m *Method) ErrorOutputs() []*Arg {
	var args []*Arg
	for _, arg := range m.Out {
		if arg.IsError() {
			args = append(args, arg)
		}
	}
	return args
}

func (m *Method) NonErrorOutputs() []*Arg {
	var args []*Arg
	for _, arg := range m.Out {
		if !arg.IsError() {
			args = append(args, arg)
		}
	}
	return args
}

type Interface struct {
	Name       string
	Methods    []*Method           `json:",omitempty"`
	Comment    string             `json:",omitempty"`
	Definition *ast.InterfaceType `json:"-"`
}

type Printer struct {
	Src        string
	CommentMap ast.CommentMap
}

func (p *Printer) ToString(node ast.Node) string {
	if node == nil {
		return ""
	}
	return p.Src[node.Pos()-1:node.End()-1]
}

func (in *Interface) Method(name string) *Method {
	for _, m := range in.Methods {
		if m.Name == name {
			return m
		}
	}
	return nil
}

type File struct {
	Package    string
	Comment    string
	Imports    map[string]string
	Values     []*Value
	Interfaces []*Interface `json:",omitempty"`
	Structs    []*Struct    `json:",omitempty"`
	Printer    *Printer    `json:"-"`
}

func (f *File) Value(name string) *Value {
	for _, v := range f.Values {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (f *File) Interface(name string) *Interface {
	for _, v := range f.Interfaces {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func Scan(filename string) (*File, error) {
	tokens := token.NewFileSet()
	file, err := parser.ParseFile(tokens, filename, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	printer := &Printer{string(content), ast.NewCommentMap(tokens, file, file.Comments)}

	var structs []*Struct
	for _, node := range file.Decls {
		structs = append(structs, Structs(printer, node)...)
	}
	var interfaces []*Interface
	for _, node := range file.Decls {
		interfaces = append(interfaces, Interfaces(printer, node)...)
	}

	var constants []*Value
	for _, node := range file.Decls {
		for _, v := range Values(printer, node) {
			constants = append(constants, v)
		}
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
		Values:     constants,
		Comment:    joinComments(printer.CommentMap[file]),
	}, nil

}

func InterfacesFile(filename string) ([]*Interface, *Printer, error) {
	tokens := token.NewFileSet()
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	file, err := parser.ParseFile(tokens, filename, nil, parser.AllErrors|parser.ParseComments)
	printer := &Printer{string(content), ast.NewCommentMap(tokens, file, file.Comments)}
	if err != nil {
		return nil, nil, err
	}
	var res []*Interface
	for _, node := range file.Decls {
		res = append(res, Interfaces(printer, node)...)
	}
	return res, printer, nil
}

func Interfaces(printer *Printer, decls ...ast.Node) []*Interface {
	var res []*Interface
	var stack []ast.Node
	var name string
	for i := len(decls) - 1; i >= 0; i-- {
		stack = append(stack, decls[i])
	}
	var lastComment string
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch v := node.(type) {
		case *ast.TypeSpec:
			name = v.Name.Name
			stack = append(stack, v.Type)
		case *ast.InterfaceType:
			iface := &Interface{Name: name, Definition: v, Comment: lastComment}
			for _, m := range v.Methods.List {
				// skip anonym method - in progress
				if len(m.Names) == 0 {
					continue
				}
				iface.Methods = append(iface.Methods, AsMethod(m, printer))
			}
			res = append(res, iface)
		case *ast.GenDecl:
			lastComment = joinComments(printer.CommentMap[v])
			for _, spec := range v.Specs {
				stack = append(stack, spec)
			}
		}
	}
	return res
}

func Values(printer *Printer, decls ...ast.Node) map[string]*Value {
	var res = make(map[string]*Value)
	var stack []ast.Node
	//var name string
	for i := len(decls) - 1; i >= 0; i-- {
		stack = append(stack, decls[i])
	}
	var lastComment string
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		switch v := node.(type) {
		case *ast.ValueSpec:
			val := &Value{}
			val.Name = v.Names[0].Name
			val.Type = v.Type
			val.Value = v.Values[0] // TODO: check for bounds
			val.printer = printer
			val.Comment = lastComment
			res[val.Name] = val
		case *ast.GenDecl:
			lastComment = joinComments(printer.CommentMap[v])
			for _, spec := range v.Specs {
				stack = append(stack, spec)
			}
		}
	}
	return res
}

func AsMethod(m *ast.Field, printer *Printer) *Method {
	var name string
	name = m.Names[0].Name
	method := &Method{Name: name, Comment: joinComments(printer.CommentMap[m])}
	def := m.Type.(*ast.FuncType)
	if def.Params != nil {
		method.In = getArgs(printer, def.Params.List)
	}
	if def.Results != nil {
		for i, p := range def.Results.List {
			name := fmt.Sprintf("ret%v", i)
			if p.Names != nil {
				name = p.Names[0].Name
			}
			method.Out = append(method.Out, &Arg{name, p.Type, joinComments(printer.CommentMap[p]), printer})
		}
	}
	return method
}

func getArgs(printer *Printer, fields []*ast.Field) []*Arg {
	var ans []*Arg
	for i, p := range fields {
		if p.Names != nil {
			for _, name := range p.Names {
				ans = append(ans, &Arg{name.Name, p.Type, joinComments(printer.CommentMap[p]), printer})
			}
		} else {
			ans = append(ans, &Arg{fmt.Sprintf("arg%v", i), p.Type, joinComments(printer.CommentMap[p]), printer})
		}
	}
	return ans
}

func joinComments(comments []*ast.CommentGroup) string {
	var ans string
	for _, c := range comments {
		if ans != "" {
			ans += "\n"
		}
		ans += c.Text()
	}
	return ans
}
