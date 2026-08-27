package t32

import "fmt"

type GIFTag struct {
	Low  uint64 `json:"low"`
	High uint64 `json:"high"`
}

func (g GIFTag) NLoop() uint16 {
	return uint16(g.Low & 0x7fff)
}

type GIFAD struct {
	Value    uint64 `json:"value"`
	Register uint64 `json:"register"`
}

type GIFUploadPacket struct {
	SetupTag  GIFTag `json:"setup_tag"`
	BITBLTBUF GIFAD  `json:"bitbltbuf"`
	TRXPOS    GIFAD  `json:"trxpos"`
	TRXREG    GIFAD  `json:"trxreg"`
	TRXDIR    GIFAD  `json:"trxdir"`
	ImageTag  GIFTag `json:"image_tag"`
}

func ValidateGIFUploadPacket(packet GIFUploadPacket) error {
	if packet.SetupTag.NLoop() != 4 {
		return fmt.Errorf("setup GIFtag has NLOOP=%d, expected 4", packet.SetupTag.NLoop())
	}
	registers := [4]uint64{
		packet.BITBLTBUF.Register,
		packet.TRXPOS.Register,
		packet.TRXREG.Register,
		packet.TRXDIR.Register,
	}
	expected := [4]uint64{
		GIFRegisterBITBLTBUF,
		GIFRegisterTRXPOS,
		GIFRegisterTRXREG,
		GIFRegisterTRXDIR,
	}
	for i := range registers {
		if registers[i] != expected[i] {
			return fmt.Errorf(
				"setup write %d targets GS register %#x, expected %#x",
				i,
				registers[i],
				expected[i],
			)
		}
	}
	return nil
}
