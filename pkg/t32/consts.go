package t32

import (
	"sync"
)

const (
	HeaderSize       = 0x20
	TextureEntrySize = 0x20
	DMATagSize       = 0x10
	UploadHeaderSize = 0x60

	GIFRegisterBITBLTBUF = 0x50
	GIFRegisterTRXPOS    = 0x51
	GIFRegisterTRXREG    = 0x52
	GIFRegisterTRXDIR    = 0x53
)

var (
	gsMutex sync.Mutex
)
