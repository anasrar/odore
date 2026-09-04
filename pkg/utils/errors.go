package utils

import "errors"

var (
	ErrNotIndexed     = errors.New("texture is not PSMT8 or PSMT4 indexed data")
	ErrContainerIsNil = errors.New("Container is nil")
	ErrStreamIsNil    = errors.New("stream is nil")
)
