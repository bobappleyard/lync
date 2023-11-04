package format

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

var (
	ErrUnknownType    = errors.New("unknown type")
	ErrNotPointer     = errors.New("not a pointer")
	ErrUnhandledValue = errors.New("unhandled value")
)

func Marshal(x any) ([]byte, error) {
	return MarshalInto(nil, x)
}

func Unmarshal(buf []byte, ref any) error {
	_, err := UnmarshalFrom(buf, ref)
	return err
}

func MarshalInto(buf []byte, x any) ([]byte, error) {
	xv := reflect.ValueOf(x)
	m := marshaler{buf: buf}
	err := handleValue(&m, xv)
	if err != nil {
		return nil, err
	}
	return m.buf, nil
}

func UnmarshalFrom(buf []byte, x any) ([]byte, error) {
	xv := reflect.ValueOf(x)
	if xv.Kind() != reflect.Pointer {
		return nil, ErrNotPointer
	}
	m := unmarshaler{buf: buf}
	err := handleValue(&m, xv.Elem())
	if err != nil {
		return nil, err
	}
	return m.buf, nil
}

type marshaler struct {
	buf []byte
}

// handleBool implements valueHandler.
func (m *marshaler) handleBool(v reflect.Value) error {
	if v.Bool() {
		m.buf = append(m.buf, 1)
	} else {
		m.buf = append(m.buf, 0)
	}
	return nil
}

// handleUvarint implements valueHandler.
func (m *marshaler) handleUvarint(v reflect.Value) error {
	m.buf = binary.AppendUvarint(m.buf, v.Uint())
	return nil
}

// handleVarint implements valueHandler.
func (m *marshaler) handleVarint(v reflect.Value) error {
	m.buf = binary.AppendVarint(m.buf, v.Int())
	return nil
}

// handleArray implements valueHandler.
func (m *marshaler) handleArray(v reflect.Value, size int) error {
	for i := 0; i < size; i++ {
		err := handleValue(m, v.Index(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// handleBytes implements valueHandler.
func (m *marshaler) handleBytes(v reflect.Value) error {
	bs := v.Bytes()
	m.buf = binary.AppendUvarint(m.buf, uint64(len(bs)))
	m.buf = append(m.buf, bs...)
	return nil
}

// handleInt implements valueHandler.
func (m *marshaler) handleInt(v reflect.Value, size int) error {
	value := v.Int()
	for i := 0; i < size; i++ {
		m.buf = append(m.buf, byte(value>>(i*8)))
	}
	return nil
}

// handleSlice implements valueHandler.
func (m *marshaler) handleSlice(v reflect.Value) error {
	n := v.Len()
	m.buf = binary.AppendUvarint(m.buf, uint64(n))
	for i := 0; i < n; i++ {
		err := handleValue(m, v.Index(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// handleString implements valueHandler.
func (m *marshaler) handleString(v reflect.Value) error {
	bs := v.String()
	m.buf = binary.AppendUvarint(m.buf, uint64(len(bs)))
	m.buf = append(m.buf, bs...)
	return nil

}

// handleStruct implements valueHandler.
func (m *marshaler) handleStruct(v reflect.Value) error {
	for _, f := range reflect.VisibleFields(v.Type()) {
		err := handleValue(m, v.FieldByIndex(f.Index))
		if err != nil {
			return err
		}
	}
	return nil
}

// handleUint implements valueHandler.
func (m *marshaler) handleUint(v reflect.Value, size int) error {
	value := v.Uint()
	for i := 0; i < size; i++ {
		m.buf = append(m.buf, byte(value>>(i*8)))
	}
	return nil
}

type unmarshaler struct {
	buf []byte
}

// handleBool implements valueHandler.
func (u *unmarshaler) handleBool(v reflect.Value) error {
	if len(u.buf) < 1 {
		return io.ErrUnexpectedEOF
	}
	v.SetBool(u.buf[0] != 0)
	u.buf = u.buf[1:]
	return nil
}

// handleUvarint implements valueHandler.
func (u *unmarshaler) handleUvarint(v reflect.Value) error {
	n, err := u.readInt()
	if err != nil {
		return err
	}
	v.SetUint(n)
	return nil
}

// handleVarint implements valueHandler.
func (u *unmarshaler) handleVarint(v reflect.Value) error {
	n, w := binary.Varint(u.buf)
	if w == 0 {
		return io.ErrUnexpectedEOF
	}
	if w < 0 {
		return ErrUnhandledValue
	}

	u.buf = u.buf[w:]
	v.SetInt(n)

	return nil
}

func (u *unmarshaler) readInt() (uint64, error) {
	n, w := binary.Uvarint(u.buf)
	if w == 0 {
		return 0, io.ErrUnexpectedEOF
	}
	if w < 0 {
		return 0, ErrUnhandledValue
	}

	u.buf = u.buf[w:]

	return n, nil
}

// handleArray implements valueHandler.
func (u *unmarshaler) handleArray(v reflect.Value, size int) error {
	for i := 0; i < size; i++ {
		err := handleValue(u, v.Index(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// handleBytes implements valueHandler.
func (u *unmarshaler) handleBytes(v reflect.Value) error {
	n, err := u.readInt()
	if err != nil {
		return err
	}

	if len(u.buf) < int(n) {
		return io.ErrUnexpectedEOF
	}

	bs := u.buf[:n]
	u.buf = u.buf[n:]
	v.SetBytes(bs)

	return nil
}

// handleInt implements valueHandler.
func (u *unmarshaler) handleInt(v reflect.Value, size int) error {
	if len(u.buf) < size {
		return io.ErrUnexpectedEOF
	}

	var value int64
	for i := 0; i < size; i++ {
		value |= int64(u.buf[i]) << (i * 8)
	}
	v.SetInt(value)
	u.buf = u.buf[size:]
	return nil
}

// handleSlice implements valueHandler.
func (u *unmarshaler) handleSlice(v reflect.Value) error {
	n, err := u.readInt()
	if err != nil {
		return err
	}

	elems := reflect.MakeSlice(v.Type(), int(n), int(n))
	for i := 0; i < int(n); i++ {
		err = handleValue(u, elems.Index(i))
		if err != nil {
			return err
		}
	}

	v.Set(elems)
	return nil
}

// handleString implements valueHandler.
func (u *unmarshaler) handleString(v reflect.Value) error {
	n, err := u.readInt()
	if err != nil {
		return err
	}

	if len(u.buf) < int(n) {
		return io.ErrUnexpectedEOF
	}

	bs := u.buf[:n]
	u.buf = u.buf[n:]
	v.SetString(string(bs))

	return nil
}

// handleStruct implements valueHandler.
func (u *unmarshaler) handleStruct(v reflect.Value) error {
	for _, f := range reflect.VisibleFields(v.Type()) {
		err := handleValue(u, v.FieldByIndex(f.Index))
		if err != nil {
			return err
		}
	}
	return nil
}

// handleUint implements valueHandler.
func (u *unmarshaler) handleUint(v reflect.Value, size int) error {
	if len(u.buf) < size {
		return io.ErrUnexpectedEOF
	}

	var value uint64
	for i := 0; i < size; i++ {
		value |= uint64(u.buf[i]) << (i * 8)
	}
	v.SetUint(value)
	u.buf = u.buf[size:]
	return nil
}

type valueHandler interface {
	handleBool(v reflect.Value) error
	handleVarint(v reflect.Value) error
	handleUvarint(v reflect.Value) error
	handleInt(v reflect.Value, size int) error
	handleUint(v reflect.Value, size int) error
	handleArray(v reflect.Value, size int) error
	handleBytes(v reflect.Value) error
	handleString(v reflect.Value) error
	handleSlice(v reflect.Value) error
	handleStruct(v reflect.Value) error
}

func handleValue(h valueHandler, xv reflect.Value) error {
	switch xv.Kind() {
	case reflect.Bool:
		return h.handleBool(xv)

	case reflect.Int:
		return h.handleVarint(xv)

	case reflect.Uint:
		return h.handleUvarint(xv)

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return h.handleInt(xv, int(xv.Type().Size()))

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return h.handleUint(xv, int(xv.Type().Size()))

	case reflect.Array:
		return h.handleArray(xv, int(xv.Type().Size()))

	case reflect.Slice:
		if xv.Type().Elem().Kind() == reflect.Uint8 {
			return h.handleBytes(xv)
		}
		return h.handleSlice(xv)

	case reflect.String:
		return h.handleString(xv)

	case reflect.Struct:
		return h.handleStruct(xv)
	}

	return fmt.Errorf("%s: %w", xv.Type(), ErrUnknownType)
}
