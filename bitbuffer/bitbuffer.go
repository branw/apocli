package bitbuffer

import (
	"errors"
	"reflect"
)

//type Endian string
//
//const (
//	LittleEndian Endian = "little"
//	BigEndian    Endian = "big"
//)

type BitBuffer struct {
	data      []byte
	bitLength uint64

	bitOffset uint64
	//endian    Endian
}

func NewBitBuffer(data []byte) *BitBuffer {
	return &BitBuffer{
		data:      data,
		bitLength: uint64(len(data)) * 8,

		bitOffset: 0,
		//endian:    LittleEndian,
	}
}

func (bb *BitBuffer) BitsAvailable() uint64 {
	return bb.bitLength - bb.bitOffset
}

func (bb *BitBuffer) validateLength(bitLength uint64) error {
	if bitLength <= 0 {
		return ErrInvalidBitLength{}
	}
	if bb.bitOffset+bitLength > bb.bitLength {
		return ErrInsufficientBitsAvailable{
			BitsNeeded:    bitLength,
			BitsAvailable: bb.BitsAvailable(),
		}
	}
	return nil
}

func (bb *BitBuffer) SkipBits(bitLength uint64) error {
	if err := bb.validateLength(bitLength); err != nil {
		return err
	}
	bb.bitOffset += bitLength
	return nil
}

func (bb *BitBuffer) Seek(bytePosition uint64) error {
	bitOffset := bytePosition * 8
	if bitOffset > bb.bitLength {
		return ErrInvalidBitPosition{}
	}
	bb.bitOffset = bitOffset
	return nil
}

func (bb *BitBuffer) ReadBits(bitLength uint64) (uint64, error) {
	if err := bb.validateLength(bitLength); err != nil {
		return 0, err
	}

	value := uint64(0)
	for i := uint64(0); i < bitLength; i++ {
		bit := (bb.data[bb.bitOffset/8] >> (bb.bitOffset % 8)) & 1
		value |= uint64(bit) << i
		bb.bitOffset++
	}
	return value, nil
}

func (bb *BitBuffer) ReadUint8() (uint8, error) {
	if err := bb.validateLength(8); err != nil {
		return 0, err
	}
	if bb.bitOffset%8 == 0 {
		value := bb.data[bb.bitOffset/8]
		bb.bitOffset += 8
		return value, nil
	}
	value, err := bb.ReadBits(8)
	return uint8(value), err
}

func (bb *BitBuffer) ReadUint16() (uint16, error) {
	if err := bb.validateLength(16); err != nil {
		return 0, err
	}
	if bb.bitOffset%8 == 0 {
		offset := bb.bitOffset / 8
		value := uint16(bb.data[offset]) | (uint16(bb.data[offset+1]) << 8)
		bb.bitOffset += 16
		return value, nil
	}
	value, err := bb.ReadBits(16)
	return uint16(value), err
}

func (bb *BitBuffer) ReadUint32() (uint32, error) {
	if err := bb.validateLength(32); err != nil {
		return 0, err
	}
	if bb.bitOffset%8 == 0 {
		offset := bb.bitOffset / 8
		value := uint32(bb.data[offset]) | (uint32(bb.data[offset+1]) << 8) | (uint32(bb.data[offset+2]) << 16) | (uint32(bb.data[offset+3]) << 24)
		bb.bitOffset += 32
		return value, nil
	}
	value, err := bb.ReadBits(32)
	return uint32(value), err
}

func (bb *BitBuffer) Unmarshal(v interface{}) error {
	value := reflect.Indirect(reflect.ValueOf(v))
	if !value.CanSet() {
		return errors.New("cannot unmarshal to non-pointer type")
	}

	if err := bb.unmarshalValue(value, ""); err != nil {
		return err
	}
	return nil
}
