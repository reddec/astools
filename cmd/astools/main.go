package astools

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/reddec/astools"
	"log"
	"os"
	"encoding/json"
	"text/template"
	"github.com/Masterminds/sprig"
	"io/ioutil"
)

func main() {
	dump := kingpin.Command("dump", "Dump source AST to JSON")
	dumpGoFile := dump.Arg("input-file", "Input .go file").Required().String()

	gen := kingpin.Command("gen", "Generate result base on template, env variables and source go file")
	genGoFile := gen.Arg("input-file", "Input .go file").Required().String()
	genTemplFile := gen.Arg("template", "Go template file. Vars: .Env and .Go").Required().String()

	switch kingpin.Parse() {
	case "dump":
		data, err := atool.Scan(*dumpGoFile)
		if err != nil {
			log.Fatal("scan:", err)
		}
		bytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			log.Fatal("marshall:", err)
		}
		os.Stdout.Write(bytes)
	case "gen":
		data, err := atool.Scan(*genGoFile)
		if err != nil {
			log.Fatal("scan:", err)
		}
		templateContent, err := ioutil.ReadFile(*genTemplFile)
		if err != nil {
			log.Fatal("read template:", err)
		}
		templ, err := template.New(*genTemplFile).Funcs(sprig.TxtFuncMap()).Parse(string(templateContent))
		if err != nil {
			log.Fatal("parse:", err)
		}
		err = templ.Execute(os.Stdout, data)
		if err != nil {
			log.Fatal("render:", err)
		}
	}
}
