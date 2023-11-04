package format

import (
	"bytes"
	"reflect"
	"testing"
)

func TestEncoding(t *testing.T) {
	for _, test := range []struct {
		name string
		buf  []byte
		val  any
	}{
		{"BoolFalse", []byte{0}, false},
		{"BoolTrue", []byte{1}, true},
		{"FixInt", []byte{12, 4, 0, 0}, int32(1036)},
		{"FixUint", []byte{12, 4, 0, 0}, uint32(1036)},
		{"VarInt", []byte{254, 8}, 575},
		{"VarUint", []byte{254, 8}, uint(1150)},
		{"Bytes", []byte{3, 0, 1, 2}, []byte{0, 1, 2}},
		{"String", []byte{5, 'h', 'e', 'l', 'l', 'o'}, "hello"},
		{"Array", []byte{1, 2, 3, 4}, [4]byte{1, 2, 3, 4}},
		{"Slice", []byte{3, 0, 0, 1, 0, 2, 0}, []uint16{0, 1, 2}},
		{"Struct", []byte{4, 10}, struct{ A, B byte }{4, 10}},
	} {
		t.Run(test.name, func(t *testing.T) {
			buf, err := Marshal(test.val)
			if err != nil {
				t.Error(err)
				return
			}
			if !bytes.Equal(buf, test.buf) {
				t.Error("byte renderings differ")
			}

			ref := reflect.New(reflect.TypeOf(test.val))
			err = Unmarshal(test.buf, ref.Interface())
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(ref.Elem().Interface(), test.val) {
				t.Errorf("got %v, expected %v", ref.Elem(), test.val)
			}
		})
	}
}
