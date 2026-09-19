package sd

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/anasrar/binarium"
	"github.com/anasrar/odore/pkg/scd"
	"github.com/anasrar/odore/pkg/utils"
)

type Header struct {
	Signature        uint32 `json:"signature"`
	SCDOffset        uint32 `json:"scd_offset"`
	CollisionOffset  uint32 `json:"collision_offset"`
	MDBOffset        uint32 `json:"mdb_offset"`
	UnknownMDBOffset uint32 `json:"unknown_mdb_offset"`
}

type Container struct {
	Offset       uint32         `json:"offset" skip:""`
	Header       Header         `json:"header"`
	SCDContainer *scd.Container `json:"scd_container"`
}

func New() *Container {
	return &Container{}
}

func (c *Container) unmarshal(stream io.ReadSeeker) error {
	if stream == nil {
		return utils.ErrStreamIsNil
	}

	fileSize, err := utils.SeekerSize(stream)
	if err != nil {
		return err
	}
	baseOffset := uint64(c.Offset)
	if baseOffset > fileSize {
		return fmt.Errorf("SDSD offset %#x is outside stream size %#x",
			baseOffset, fileSize)
	}

	if err := utils.SeekAbsolute(stream, baseOffset); err != nil {
		return err
	}
	if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, &c.Header); err != nil {
		return err
	}

	if Signature != c.Header.Signature {
		return utils.ErrSignatureIsNotMatch(Signature, c.Header.Signature)
	}

	if c.Header.SCDOffset != 0 {
		scdContainer := scd.New()
		if err := scd.FromStreamWithOffset(scdContainer, stream, uint32(baseOffset)+c.Header.SCDOffset); err != nil {
			return err
		}
		c.SCDContainer = scdContainer
	}

	return nil
}

func FromStreamWithOffset(c *Container, stream io.ReadSeeker, offset uint32) error {
	if c == nil {
		return utils.ErrContainerIsNil
	}
	c.Offset = offset
	return c.unmarshal(stream)
}

func FromStream(c *Container, stream io.ReadSeeker) error {
	return FromStreamWithOffset(c, stream, 0)
}
