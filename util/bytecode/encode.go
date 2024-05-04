package main

import (
	"io"
	"strings"
	"text/template"

	"github.com/bobappleyard/lync/util/bytecode/model"
)

var encodeT = template.Must(template.New("encode").Parse(strings.TrimSpace(`

{{range .Imports}}
import {{if .Rename}}{{.Name}} {{end}}{{.Path | printf "%q" }}
{{end}}

type {{.EncoderName}} struct {
	Buf []byte
}

{{range .Opcodes}}
func (e *{{$.EncoderName}}) {{.Name}}({{range $ix, $el := .Args}}
	{{- if $ix}}, {{end -}}
	{{.Name}} {{if .Pkg}}{{.Pkg}}.{{end}}{{.Type}}{{end}}) error {
	after, err := {{$.FormatPkg}}.MarshalInto(e.Buf, uint({{.ID}}))
	if err != nil {
		return err
	}
	{{range .Args}}
	after, err = {{$.FormatPkg}}.MarshalInto(after, {{.Name}})
	if err != nil {
		return err
	}
	{{end}}
	e.Buf = after
	return nil
}
{{end}}

`)))

type EncoderScope struct {
	model.Interpreter
	EncoderName string
	FormatPkg   string
}

func encoder(w io.Writer, scope EncoderScope) error {
	return encodeT.Execute(w, scope)
}
