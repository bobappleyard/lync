package main

import (
	"io"
	"strings"
	"text/template"

	"github.com/bobappleyard/lync/util/bytecode/model"
)

var decodeT = template.Must(new(template.Template).Parse(strings.TrimSpace(`
type {{.DecoderName}} struct {
	Code []byte
	Pos int
	Impl {{.TypeName}}
}

func (d *{{.DecoderName}}) Step() (err error) {
	switch d.Code[d.Pos] {
	{{range .Opcodes}}
	case {{.ID}}:
		b := d.Code[d.Pos+1:]
		{{range .Args}}
		var {{.Name}} {{if .Pkg}}{{.Pkg}}.{{end}}{{.Type}}
		if b, err = format.UnmarshalFrom(b, &{{.Name}}); err != nil {
			return err
		}
		{{end}}
		d.Pos += len(d.Code) - len(b)
		{{if .HasError}}err = {{end}}d.Impl.{{.Name}}({{range .Args}}{{.Name}},{{end -}})
	{{end}}
	default:
		panic("unknown bytecode")
	}

	return nil
}

`)))

type DecoderScope struct {
	model.Interpreter
	DecoderName string
	TypeName    string
}

func decoder(w io.Writer, s DecoderScope) error {
	return decodeT.Execute(w, s)
}
