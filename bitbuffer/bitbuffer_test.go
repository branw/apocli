package bitbuffer

import (
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

	t.Run("", func(t *testing.T) {
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

	t.Run("", func(t *testing.T) {
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

	t.Run("", func(t *testing.T) {
		data := []uint8{0, 1}

		var s StructWithUnmarshaler
		err := Unmarshal(data, &s)
		assert.Equal(t, nil, err)
		assert.Equal(t, StructWithUnmarshaler{A: false, B: true}, s)
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
