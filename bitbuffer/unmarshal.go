package bitbuffer

import (
	"errors"
	"fmt"
	"reflect"
)

func Unmarshal(data []byte, v interface{}) error {
	bb := NewBitBuffer(data)
	err := bb.Unmarshal(v)
	if err != nil {
		return err
	}

	return nil
}

func UnmarshalFully(data []byte, v interface{}) error {
	bb := NewBitBuffer(data)
	err := bb.Unmarshal(v)
	if err != nil {
		return err
	}

	if bitsAvailable := bb.BitsAvailable(); bitsAvailable > 0 {
		return ErrBitsRemaining{
			BitsRemaining:    bitsAvailable,
			UnmarshalledType: reflect.TypeOf(v),
		}
	}
	return nil
}

type Unmarshaler interface {
	Unmarshal(bb *BitBuffer, value reflect.Value, tag reflect.StructTag) error
}

func (bb *BitBuffer) unmarshalValue(value reflect.Value, tag reflect.StructTag) error {
	if unmarshaler, ok := value.Addr().Interface().(Unmarshaler); ok {
		return unmarshaler.Unmarshal(bb, value, tag)
	}
	switch value.Kind() {
	case reflect.Struct:
		return bb.unmarshalStruct(value, tag)

	case reflect.Bool:
		return bb.unmarshalBool(value, tag)

	case reflect.Uint8:
		return bb.unmarshalUint8(value, tag)

	case reflect.Uint16:
		return bb.unmarshalUint16(value, tag)

	case reflect.Uint32:
		return bb.unmarshalUint32(value, tag)

	case reflect.Uint64:
		return bb.unmarshalUint64(value, tag)

	default:
		return errors.New(fmt.Sprintf("unknown type %+v", value.Kind()))
	}
}

func (bb *BitBuffer) unmarshalStruct(structValue reflect.Value, _ reflect.StructTag) error {
	for i := 0; i < structValue.NumField(); i++ {
		fieldValue := structValue.Field(i)
		fieldTag := structValue.Type().Field(i).Tag
		if err := bb.unmarshalValue(fieldValue, fieldTag); err != nil {
			return err
		}
	}

	return nil
}

func (bb *BitBuffer) unmarshalBool(boolValue reflect.Value, tag reflect.StructTag) error {
	parsedTag, err := ParseTag(tag)
	if err != nil {
		return err
	}
	readValue, err := bb.ReadBits(parsedTag.BitWidthOrDefault(1))
	if err != nil {
		return err
	}
	boolValue.SetBool(readValue != 0)
	return nil
}

func (bb *BitBuffer) unmarshalUint8(uint8Value reflect.Value, tag reflect.StructTag) error {
	parsedTag, err := ParseTag(tag)
	if err != nil {
		return err
	}
	readValue, err := bb.ReadBits(parsedTag.BitWidthOrDefault(8))
	if err != nil {
		return err
	}
	uint8Value.SetUint(readValue)
	return nil
}

func (bb *BitBuffer) unmarshalUint16(uint16Value reflect.Value, tag reflect.StructTag) error {
	parsedTag, err := ParseTag(tag)
	if err != nil {
		return err
	}
	readValue, err := bb.ReadBits(parsedTag.BitWidthOrDefault(16))
	if err != nil {
		return err
	}
	uint16Value.SetUint(readValue)
	return nil
}

func (bb *BitBuffer) unmarshalUint32(uint32Value reflect.Value, tag reflect.StructTag) error {
	parsedTag, err := ParseTag(tag)
	if err != nil {
		return err
	}
	readValue, err := bb.ReadBits(parsedTag.BitWidthOrDefault(32))
	if err != nil {
		return err
	}
	uint32Value.SetUint(readValue)
	return nil
}

func (bb *BitBuffer) unmarshalUint64(uint64Value reflect.Value, tag reflect.StructTag) error {
	parsedTag, err := ParseTag(tag)
	if err != nil {
		return err
	}
	readValue, err := bb.ReadBits(parsedTag.BitWidthOrDefault(64))
	if err != nil {
		return err
	}
	uint64Value.SetUint(readValue)
	return nil
}
