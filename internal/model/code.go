package model

import (
	"errors"
	"math/rand/v2"
	"strconv"
)

var ErrInvalidConfirmationCodeLength = errors.New("confirmation code must consist of 4 digits")

type Code struct{ value int }

func (c Code) String() string {
	return strconv.Itoa(c.value)
}
func (c Code) Bytes() []byte {
	return []byte(c.String())
}

func NewCode() Code {
	return Code{rand.IntN(9000) + 1000}
}

func NewCodeFromInt(codeInt int) (Code, error) {
	if codeInt < 1000 || codeInt > 9999 {
		return Code{}, ErrInvalidConfirmationCodeLength
	}
	return Code{codeInt}, nil
}
