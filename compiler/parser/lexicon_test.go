package parser

import (
	"testing"

	"github.com/bobappleyard/lync/util/assert"
)

func TestLex(t *testing.T) {
	s, err := tokenize([]byte(`

{
	var aVariable = object.method("an argument\n\"with quotes in\"", 123)
}

	`))
	assert.Nil(t, err)
	assert.Equal(t, s, []token{
		tokenType[newlineTok](0, "\n\n"),
		tokenType[openBTok](2, "{"),
		tokenType[newlineTok](3, "\n\t"),
		tokenType[varTok](5, "var"),
		tokenType[idTok](9, "aVariable"),
		tokenType[eqTok](19, "="),
		tokenType[idTok](21, "object"),
		tokenType[dotTok](27, "."),
		tokenType[idTok](28, "method"),
		tokenType[openPTok](34, "("),
		tokenType[stringTok](35, `"an argument\n\"with quotes in\""`),
		tokenType[commaTok](68, ","),
		tokenType[intTok](70, "123"),
		tokenType[closePTok](73, ")"),
		tokenType[newlineTok](74, "\n"),
		tokenType[closeBTok](75, "}"),
		tokenType[newlineTok](76, "\n\n\t"),
	})

	t.Logf("%#v", s)
}
