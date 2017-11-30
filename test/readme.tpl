{{range .Interfaces }}

{{.Comment}}

Path: {{ env "PATH"}}

## {{$.Package}}.{{.Name}}
{{end}}