package main

import (
	"flag"
	"fmt"
	"go/types"
	"os"
	"slices"
	"strings"

	"github.com/bobappleyard/lync/util/bytecode/model"
	"golang.org/x/tools/go/packages"
)

var (
	src = flag.String("src", "", "")
	enc = flag.String("enc", "", "")
	dec = flag.String("dec", "", "")
	out = flag.String("out", "gen_bytecode.go", "")
)

func main() {
	flag.Parse()

	pkg := loadPackage(".")
	typ := pkg.Scope().Lookup(*src)
	if typ == nil {
		fmt.Println("unkown type")
		return
	}

	interp := parseInterpreter(typ.(*types.TypeName).Type().(*types.Named))
	formatPkg := packageName(&interp, loadPackage("github.com/bobappleyard/lync/util/format"))

	f, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString(`// Code generated by github.com/bobappleyard/lync/util/bytecode DO NOT EDIT
package `)
	f.WriteString(pkg.Name())

	encoderName := *enc
	if encoderName == "" {
		encoderName = *src + "Encoder"
	}
	if err := encoder(f, EncoderScope{
		Interpreter: interp,
		EncoderName: encoderName,
		FormatPkg:   formatPkg,
	}); err != nil {
		panic(err)
	}

	decoderName := *dec
	if decoderName == "" {
		decoderName = *src + "Decoder"
	}
	if err := decoder(f, DecoderScope{
		Interpreter: interp,
		DecoderName: decoderName,
		TypeName:    *src,
	}); err != nil {
		panic(err)
	}
}

func loadPackage(name string) *types.Package {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes,
	}, name)
	if err != nil {
		panic(err)
	}
	return pkgs[0].Types
}

func parseInterpreter(typ *types.Named) model.Interpreter {
	var interp model.Interpreter
	for i := 0; i < typ.NumMethods(); i++ {
		m := typ.Method(i)
		if !m.Exported() {
			continue
		}
		t := m.Type().(*types.Signature)

		op := model.Opcode{
			ID:   i,
			Name: m.Name(),
		}

		for j := 0; j < t.Params().Len(); j++ {
			param := t.Params().At(j)
			var pkg, typeName string
			if t, ok := param.Type().(*types.Named); ok {
				pkg = packageName(&interp, t.Obj().Pkg())
				typeName = t.Obj().Name()
			} else {
				typeName = param.Type().String()
			}

			op.Args = append(op.Args, model.Arg{
				Name: param.Name(),
				Pkg:  pkg,
				Type: typeName,
			})
		}
		interp.Opcodes = append(interp.Opcodes, op)
	}
	return interp
}

func packageName(interp *model.Interpreter, pkg *types.Package) string {
	idx, ok := slices.BinarySearchFunc(
		interp.Imports,
		pkg.Path(),
		func(i model.Import, s string) int {
			return strings.Compare(s, i.Path)
		},
	)
	if ok {
		return interp.Imports[idx].Name
	}

	name, rename := importName(interp, pkg.Name())

	interp.Imports = append(interp.Imports, model.Import{})
	copy(interp.Imports[idx+1:], interp.Imports[idx:])
	interp.Imports[idx] = model.Import{
		Path:   pkg.Path(),
		Name:   name,
		Rename: rename,
	}

	return name
}

func importName(interp *model.Interpreter, name string) (string, bool) {
	dupes := 0
	test := func(i model.Import) bool {
		return i.Name == name
	}
	for slices.ContainsFunc(interp.Imports, test) {
		dupes++
		name = fmt.Sprintf("%s%d", name, dupes)
	}
	return name, dupes != 0
}