package transform

import (
	"testing"

	"github.com/bobappleyard/lync/util/assert"
)

func TestVars(t *testing.T) {

	vs := vars{}

	vs.add("a")
	vs.add("b")

	assert.Equal(t, len(vs), 2)

	vs.add("d")
	vs.add("c")
	vs.add("a")

	assert.Equal(t, len(vs), 4)

	assert.True(t, vs.contains("a"))
	assert.False(t, vs.contains("x"))
}
