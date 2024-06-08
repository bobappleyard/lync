package assert

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/r3labs/diff"
)

func Equal[T any](t testing.TB, got, expected T) {
	t.Helper()
	if !reflect.DeepEqual(got, expected) {
		cl, err := diff.Diff(expected, got)
		if err != nil {
			t.Error(err)
		}
		if cl != nil {
			sb, err := json.MarshalIndent(cl, "", "\t")
			if err != nil {
				panic(err)
			}
			t.Errorf("Mismatched values\nDiff:\n%s", string(sb))
		}
	}
}

func Nil(t testing.TB, got any) {
	t.Helper()
	if got == nil {
		return
	}
	if reflect.ValueOf(got).IsNil() {
		return
	}
	t.Errorf("got %v, expecting nil", got)
}

func True(t testing.TB, got bool) {
	t.Helper()
	Equal(t, got, true)
}

func False(t testing.TB, got bool) {
	t.Helper()
	Equal(t, got, false)
}
