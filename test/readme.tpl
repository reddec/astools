{{range .Interfaces }}

Path: {{ env "PATH"}}

## {{$.Package}}.{{.Name}}
{{end}}