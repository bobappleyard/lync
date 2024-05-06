package parser

import (
	"strconv"
	"strings"
	"unsafe"

	"github.com/bobappleyard/lync/util/must"
	"github.com/bobappleyard/lync/util/text"
)

type token interface {
	start() int
	text() string
}

type tokenData struct {
	s int
	t string
}

func (d tokenData) start() int {
	return d.s
}

func (d tokenData) text() string {
	return d.t
}

func tokenType[T token](start int, text string) token {
	var res T
	data := (*tokenData)(unsafe.Pointer(&res))
	data.s = start
	data.t = text
	return res
}

// basic elements

type stringTok struct{ tokenData }
type intTok struct{ tokenData }
type fltTok struct{ tokenData }
type idTok struct{ tokenData }

func (t stringTok) value() string {
	v, _ := strconv.Unquote(t.text())
	return v
}

func (t intTok) value() int {
	v, _ := strconv.Atoi(t.text())
	return v
}

func (t fltTok) value() float64 {
	v, _ := strconv.ParseFloat(t.text(), 64)
	return v
}

// punctuation

type eqTok struct{ tokenData }
type dotTok struct{ tokenData }
type commaTok struct{ tokenData }
type openPTok struct{ tokenData }
type closePTok struct{ tokenData }
type openBTok struct{ tokenData }
type closeBTok struct{ tokenData }
type spaceTok struct{ tokenData }
type newlineTok struct{ tokenData }

// keywords

type varTok struct{ tokenData }
type classTok struct{ tokenData }
type funcTok struct{ tokenData }
type ifTok struct{ tokenData }
type importTok struct{ tokenData }
type returnTok struct{ tokenData }

var keywords = map[string]text.TokenConstructor[token]{
	"var":    tokenType[varTok],
	"class":  tokenType[classTok],
	"func":   tokenType[funcTok],
	"if":     tokenType[ifTok],
	"import": tokenType[importTok],
	"return": tokenType[returnTok],
}

func tokenize(src []byte) ([]token, error) {
	toks, err := lexer.Tokenize(src).Force()
	if err != nil {
		return nil, err
	}
	var res []token
	var context []token
	for _, t := range toks {
		switch t := t.(type) {
		case spaceTok:
			if tokenIsNewline(t, context) {
				res = append(res, newlineTok(t))
			}
		case openBTok, openPTok:
			context = append(context, t)
			res = append(res, t)
		case closePTok, closeBTok:
			context = context[:len(context)-1]
			res = append(res, t)
		default:
			res = append(res, t)
		}
	}
	return res, nil
}

func tokenIsNewline(t token, context []token) bool {
	n := len(context)
	return (n == 0 || context[n-1].text() != "(") &&
		strings.Contains(t.text(), "\n")
}

func tokenIdType(start int, text string) token {
	if kw, ok := keywords[text]; ok {
		return kw(start, text)
	}
	return tokenType[idTok](start, text)
}

var lexer = must.Be(text.NewLexer(
	text.Regex(`"([^"]|\\.)*"`, tokenType[stringTok]),
	text.Regex(`\d+`, tokenType[intTok]),
	text.Regex(`\d+\.\d+`, tokenType[fltTok]),
	text.Regex(`[a-zA-Z_]\w*`, tokenIdType),
	text.Regex(`=`, tokenType[eqTok]),
	text.Regex(`\s+`, tokenType[spaceTok]),
	text.Regex(`\.`, tokenType[dotTok]),
	text.Regex(`,`, tokenType[commaTok]),
	text.Regex(`\(`, tokenType[openPTok]),
	text.Regex(`\)`, tokenType[closePTok]),
	text.Regex(`{`, tokenType[openBTok]),
	text.Regex(`}`, tokenType[closeBTok]),
))
