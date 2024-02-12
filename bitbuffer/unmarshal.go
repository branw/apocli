package bitbuffer

import (
	"errors"
	"fmt"
	"reflect"
)

type Unmarshaler interface {
	Unmarshal(bb *BitBuffer, value reflect.Value, tag reflect.StructTag) error
}

type Validatable interface {
	Validate() error
}

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

func (bb *BitBuffer) unmarshalValue(value reflect.Value, tag reflect.StructTag) error {
	if unmarshaler, ok := value.Addr().Interface().(Unmarshaler); ok {
		return unmarshaler.Unmarshal(bb, value, tag)
	}

	parsedTag, err := ParseTag(tag)
	if err != nil {
		return err
	}

	switch value.Kind() {
	case reflect.Struct:
		err = bb.unmarshalStruct(value, parsedTag)
	case reflect.Bool:
		err = bb.unmarshalBool(value, parsedTag)
	case reflect.Uint8:
		err = bb.unmarshalUint8(value, parsedTag)
	case reflect.Uint16:
		err = bb.unmarshalUint16(value, parsedTag)
	case reflect.Uint32:
		err = bb.unmarshalUint32(value, parsedTag)
	case reflect.Uint64:
		err = bb.unmarshalUint64(value, parsedTag)

	default:
		return errors.New(fmt.Sprintf("unknown type %+v", value.Kind()))
	}
	if err != nil {
		return err
	}

	if parsedTag.Validate {
		if err = validate(value); err != nil {
			return err
		}
	}

	return nil
}

func validate(value reflect.Value) error {
	if validator, ok := value.Addr().Interface().(Validatable); ok {
		err := validator.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (bb *BitBuffer) unmarshalStruct(structValue reflect.Value, _ ParsedTag) error {
	for i := 0; i < structValue.NumField(); i++ {
		fieldValue := structValue.Field(i)
		fieldTag := structValue.Type().Field(i).Tag
		if err := bb.unmarshalValue(fieldValue, fieldTag); err != nil {
			return err
		}
	}
	return nil
}

func (bb *BitBuffer) unmarshalBool(boolValue reflect.Value, parsedTag ParsedTag) error {
	readValue, err := bb.ReadBits(parsedTag.BitWidthOrDefault(1))
	if err != nil {
		return err
	}
	boolValue.SetBool(readValue != 0)
	return nil
}

func (bb *BitBuffer) unmarshalUint8(uint8Value reflect.Value, parsedTag ParsedTag) error {
	readValue, err := bb.ReadBits(parsedTag.BitWidthOrDefault(8))
	if err != nil {
		return err
	}
	uint8Value.SetUint(readValue)
	return nil
}

func (bb *BitBuffer) unmarshalUint16(uint16Value reflect.Value, parsedTag ParsedTag) error {
	readValue, err := bb.ReadBits(parsedTag.BitWidthOrDefault(16))
	if err != nil {
		return err
	}
	uint16Value.SetUint(readValue)
	return nil
}

func (bb *BitBuffer) unmarshalUint32(uint32Value reflect.Value, parsedTag ParsedTag) error {
	readValue, err := bb.ReadBits(parsedTag.BitWidthOrDefault(32))
	if err != nil {
		return err
	}
	uint32Value.SetUint(readValue)
	return nil
}

func (bb *BitBuffer) unmarshalUint64(uint64Value reflect.Value, parsedTag ParsedTag) error {
	readValue, err := bb.ReadBits(parsedTag.BitWidthOrDefault(64))
	if err != nil {
		return err
	}
	uint64Value.SetUint(readValue)
	return nil
}
