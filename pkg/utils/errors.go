package utils

import (
	"errors"
	"fmt"
)

var (
	ErrNotIndexed     = errors.New("texture is not PSMT8 or PSMT4 indexed data")
	ErrContainerIsNil = errors.New("Container is nil")
	ErrStreamIsNil    = errors.New("stream is nil")
)

func ErrSignatureIsNotMatch(expect, actual uint32) error {
	return fmt.Errorf("signature is not match: expect %#x, actual: %#x", expect, actual)
}
