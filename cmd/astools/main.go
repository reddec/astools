package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/reddec/astools"
	"log"
	"os"
	"encoding/json"
	"text/template"
	"github.com/Masterminds/sprig"
	"io/ioutil"
	"path/filepath"
	"io"
	"bytes"
	"strings"
	"path"
)

func main() {
	dump := kingpin.Command("dump", "Dump source AST to JSON")
	dumpGoFile := dump.Arg("input-file", "Input .go file").Required().String()

	gen := kingpin.Command("gen", "Generate result base on template, env variables and source go file")
	genGoFile := gen.Arg("input-file", "Input .go file").Required().String()
	genTemplFile := gen.Arg("template", "Go template file. Vars: .Env and .Go").Required().Strings()
	genExt := gen.Flag("ext", "Remove extension for output files").Short('e').Bool()
	genOutput := gen.Flag("out", "Output folder. If not specified - to stdout").Short('o').String()
	genCopy := gen.Flag("copy", "Copy original file to output (if specified)").Short('c').Bool()

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
		templates := template.New("").Funcs(sprig.TxtFuncMap())
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
				err = ioutil.WriteFile(target, (out.(*bytes.Buffer)).Bytes(), 755)
				if err != nil {
					log.Fatal("save to", target, ":", err)
				}
			}
		}
		if *genOutput != "" && *genCopy {
			target := path.Join(*genOutput, filepath.Base(*genGoFile))
			err = ioutil.WriteFile(target, []byte(data.Printer.Src), 755)
			if err != nil {
				log.Fatal("copy to", target, ":", err)
			}
		}
	}
}
