package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/reddec/astools"
	"log"
	"os"
	"encoding/json"
	"text/template"
	"strings"
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
		templ, err := template.ParseFiles(*genTemplFile)
		if err != nil {
			log.Fatal("parse:", err)
		}
		var env = make(map[string]string)
		for _, item := range os.Environ() {
			kv := strings.SplitN(item, "=", 2)
			if len(kv) == 2 {
				env[kv[0]] = kv[1]
			}
		}
		err = templ.Execute(os.Stdout, map[string]interface{}{
			"Env": env,
			"Go":  data,
		})
		if err != nil {
			log.Fatal("render:", err)
		}
	}
}
