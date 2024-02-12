package bitbuffer

import (
	"fmt"
	"reflect"
)

type ErrInvalidBitLength struct{}

func (err ErrInvalidBitLength) Error() string {
	return "invalid bit length"
}

type ErrInvalidBitPosition struct{}

func (err ErrInvalidBitPosition) Error() string {
	return "invalid bit position"
}

type ErrInsufficientBitsAvailable struct {
	BitsNeeded    uint64
	BitsAvailable uint64
}

func (err ErrInsufficientBitsAvailable) Error() string {
	return fmt.Sprintf("insufficient bits available (need %d, have %d)",
		err.BitsNeeded, err.BitsAvailable)
}

type ErrBitsRemaining struct {
	BitsRemaining    uint64
	UnmarshalledType reflect.Type
}

func (err ErrBitsRemaining) Error() string {
	return fmt.Sprintf("%d bits remaining in bitbuffer after unmarshal", err.BitsRemaining)
}
