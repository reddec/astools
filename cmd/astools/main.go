package main

import (
	"bytes"
	"encoding/json"
	"github.com/Masterminds/sprig"
	"github.com/reddec/astools"
	"github.com/reddec/symbols"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	dump := kingpin.Command("dump", "Dump source AST to JSON")
	dumpFilter := dump.Flag("filter", "Filter output (used name flag)").Short('f').Default("all").Enum("all", "struct", "interface", "value")
	dumpFilterName := dump.Flag("filter-name", "Filter name").Short('n').String()
	dumpGoFile := dump.Arg("input-file", "Input .go file").Required().String()

	gen := kingpin.Command("gen", "Generate result base on template, env variables and source go file")
	genGoFile := gen.Arg("input-file", "Input .go file").Required().String()
	genTemplFile := gen.Arg("template", "Go template file. Vars: .Env and .Go").Required().Strings()
	genExt := gen.Flag("ext", "Remove extension for output files").Short('e').Bool()
	genOutput := gen.Flag("out", "Output folder. If not specified - to stdout").Short('o').String()
	genCopy := gen.Flag("copy", "Copy original file to output (if specified)").Short('c').Bool()
	indexSymbols := gen.Flag("index", "Index all symbols during generations (sym func)").Short('I').Bool()

	switch kingpin.Parse() {
	case "dump":
		data, err := atool.Scan(*dumpGoFile)
		if err != nil {
			log.Fatal("scan:", err)
		}

		var res interface{} = data

		switch *dumpFilter {
		case "struct":
			res = data.Struct(*dumpFilterName)
		case "interface":
			res = data.Interface(*dumpFilterName)
		case "value":
			res = data.Value(*dumpFilterName)
		case "all":
			res = data
		default:
			log.Fatal("unknown filter mode:", *dumpFilter)
		}

		dump, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			log.Fatal("marshall:", err)
		}
		os.Stdout.Write(dump)
	case "gen":
		data, err := atool.Scan(*genGoFile)
		if err != nil {
			log.Fatal("scan:", err)
		}
		funcs := sprig.TxtFuncMap()
		if *indexSymbols {
			project, err := symbols.ProjectByDir(filepath.Dir(*genGoFile), 8192)
			if err != nil {
				log.Fatal("index symbols: ", err)
			}
			var currentFile *symbols.File
			for _, f := range project.Package.Files {
				if filepath.Base(f.Filename) == filepath.Base(*genGoFile) {
					currentFile = f
					break
				}
			}
			if currentFile == nil {
				panic("indexed but not found")
			}
			funcs["sym"] = func() *symbols.Project {
				return project
			}
			funcs["symfile"] = func() *symbols.File {
				return currentFile
			}
			funcs["fqdn"] = func(arg *atool.Arg) (string, error) {
				v, err := project.FindSymbol(arg.GolangType(), currentFile)
				if err != nil {
					return "", err
				}
				if arg.IsPointer() {
					return "*" + v.Import.Package + "." + v.Name, nil
				}
				return v.Import.Package + "." + v.Name, nil
			}
			funcs["symbol"] = func(name string) (*symbols.Symbol, error) {
				return project.FindLocalSymbol(name)
			}
			funcs["fields"] = func(name string) ([]*symbols.Field, error) {
				sym, err := project.FindLocalSymbol(name)
				if err != nil {
					return nil, err
				}
				return sym.Fields(project)
			}
		}
		templates := template.New("").Funcs(funcs)
		for _, fileName := range *genTemplFile {
			templateContent, err := ioutil.ReadFile(fileName)
			if err != nil {
				log.Fatal("read template:", err)
			}
			_, err = templates.New(fileName).Parse(string(templateContent))
			if err != nil {
				log.Fatal("parse:", err)
			}
		}
		if *genOutput != "" {
			err := os.MkdirAll(*genOutput, 0755)
			if err != nil {
				log.Fatal("create output dir:", err)
			}
		}
		for _, fileName := range *genTemplFile {
			var out io.Writer = os.Stdout
			if *genOutput != "" {
				out = &bytes.Buffer{}
			}
			err = templates.ExecuteTemplate(out, fileName, data)
			if err != nil {
				log.Fatal("render:", err)
			}
			if *genOutput != "" {
				if *genExt {
					idx := strings.LastIndex(fileName, ".")
					if idx > 0 {
						fileName = fileName[:idx]
					}
				}
				target := path.Join(*genOutput, filepath.Base(fileName))
				err = ioutil.WriteFile(target, (out.(*bytes.Buffer)).Bytes(), 0755)
				if err != nil {
					log.Fatal("save to", target, ":", err)
				}
			}
		}
		if *genOutput != "" && *genCopy {
			target := path.Join(*genOutput, filepath.Base(*genGoFile))
			err = ioutil.WriteFile(target, []byte(data.Printer.Src), 0755)
			if err != nil {
				log.Fatal("copy to", target, ":", err)
			}
		}
	}
}
