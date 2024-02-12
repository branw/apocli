package bitbuffer

import (
	"reflect"
	"strconv"
)

const (
	TagBitWidth = "bbwidth"
)

type ParsedTag struct {
	BitWidth uint64
}

func ParseTag(tag reflect.StructTag) (ParsedTag, error) {
	bitWidth, err := ParseBitWidthTag(tag)
	if err != nil {
		return ParsedTag{}, err
	}

	tags := ParsedTag{
		BitWidth: bitWidth,
	}
	return tags, nil
}

func (tags ParsedTag) BitWidthOrDefault(defaultBitWidth uint64) uint64 {
	if tags.BitWidth > 0 {
		return tags.BitWidth
	}
	return defaultBitWidth
}

func ParseBitWidthTag(tag reflect.StructTag) (uint64, error) {
	value, present := tag.Lookup(TagBitWidth)
	if !present {
		return 0, nil
	}

	width, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		return 0, err
	}
	return width, nil
}
