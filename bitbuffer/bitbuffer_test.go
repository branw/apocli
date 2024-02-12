package bitbuffer

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test(t *testing.T) {
	t.Run("", func(t *testing.T) {
		bb := NewBitBuffer([]uint8{0x45, 0x54})

		x, _ := bb.ReadUint8()
		y, _ := bb.ReadUint8()
		assert.Equal(t, uint8(0x45), x)
		assert.Equal(t, uint8(0x54), y)
	})

	t.Run("", func(t *testing.T) {
		bb := NewBitBuffer([]uint8{0b1111_1111})

		x, _ := bb.ReadBits(3)
		y, _ := bb.ReadBits(1)
		z, _ := bb.ReadBits(4)
		assert.Equal(t, uint64(0b111), x)
		assert.Equal(t, uint64(0b1), y)
		assert.Equal(t, uint64(0b1111), z)
	})

	t.Run("", func(t *testing.T) {
		bb := NewBitBuffer([]uint8{0b0100_0101})

		x, _ := bb.ReadBits(3)
		y, _ := bb.ReadBits(1)
		z, _ := bb.ReadBits(4)
		assert.Equal(t, uint64(0b101), x)
		assert.Equal(t, uint64(0b0), y)
		assert.Equal(t, uint64(0b0100), z)
	})
}

type StructWithUnmarshaler struct {
	A bool
	B bool
}

func (foo *StructWithUnmarshaler) Unmarshal(bb *BitBuffer, _ reflect.Value, _ reflect.StructTag) error {
	a, err := bb.ReadUint8()
	if err != nil {
		return err
	}
	b, err := bb.ReadUint8()
	if err != nil {
		return err
	}
	foo.A = a != 0
	foo.B = b != 0
	return nil
}

func TestBitBuffer_Unmarshal(t *testing.T) {
	t.Run("unmarshalling 2 bools succeeds", func(t *testing.T) {
		type Foo struct {
			A bool
			B bool
		}

		data := []uint8{0b10}

		var foo Foo
		err := Unmarshal(data, &foo)
		assert.Equal(t, nil, err)
		assert.Equal(t, Foo{A: false, B: true}, foo)
	})

	t.Run("unmarshalling 2 bools with overridden widths succeeds", func(t *testing.T) {
		type Foo struct {
			A bool `bbwidth:"8"`
			B bool `bbwidth:"8"`
		}

		data := []uint8{0, 1}

		var foo Foo
		err := Unmarshal(data, &foo)
		assert.Equal(t, nil, err)
		assert.Equal(t, Foo{A: false, B: true}, foo)
	})

	t.Run("unmarshalling a field with custom unmarshaler succeeds", func(t *testing.T) {
		data := []uint8{0, 1}

		var s StructWithUnmarshaler
		err := Unmarshal(data, &s)
		assert.Equal(t, nil, err)
		assert.Equal(t, StructWithUnmarshaler{A: false, B: true}, s)
	})

	t.Run("fully unmarshalling a struct withe exact data succeeds", func(t *testing.T) {
		type Foo struct {
			A uint8
			B uint8
		}

		data := []uint8{0x55, 0x90}

		var foo Foo
		err := UnmarshalFully(data, &foo)
		assert.Equal(t, nil, err)
		assert.Equal(t, Foo{A: 0x55, B: 0x90}, foo)
	})

	t.Run("fully unmarshalling a struct with extra data fails", func(t *testing.T) {
		type Foo struct {
			A uint8
			B uint8
		}

		data := []uint8{0x55, 0x90, 0x88}

		var foo Foo
		err := UnmarshalFully(data, &foo)
		assert.EqualError(t, err, "8 bits remaining in bitbuffer after unmarshal")
	})

	t.Run("values that are not byte-aligned cannot be unmarshalled fully", func(t *testing.T) {
		type Foo struct {
			A bool
			B bool
		}

		data := []uint8{0b10}

		var foo Foo
		err := UnmarshalFully(data, &foo)
		assert.EqualError(t, err, "6 bits remaining in bitbuffer after unmarshal")
	})
}

type ValidatedField uint8

func (vf ValidatedField) Validate() error {
	if vf > 8 {
		return errors.New("invalid")
	}
	return nil
}

type StructWithValidatedField struct {
	A uint8
	B ValidatedField `bbvalidate:"true"`
}

func TestBitBuffer_UnmarshalValidation(t *testing.T) {
	t.Run("parsing a validated field with valid data succeeds", func(t *testing.T) {
		data := []uint8{12, 6}

		var s StructWithValidatedField
		err := Unmarshal(data, &s)
		assert.Equal(t, nil, err)
		assert.Equal(t, StructWithValidatedField{A: 12, B: 6}, s)
	})

	t.Run("parsing a validated field with invalid data fails", func(t *testing.T) {
		data := []uint8{12, 14}

		var s StructWithValidatedField
		err := Unmarshal(data, &s)
		assert.EqualError(t, err, "invalid")
	})
}
