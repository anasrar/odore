package ppk

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/anasrar/binarium"
	"github.com/anasrar/odore/pkg/utils"
)

type RegionAddressSize struct {
	Offset uint32 `json:"offset"`
	Size   uint32 `json:"size"`
}

type Header struct {
	Signature uint32 `json:"signature"`
	Unknown0  uint32 `json:"unknown0"`
}

type Regions struct {
	Region0 RegionAddressSize `json:"region0"`
	Region1 RegionAddressSize `json:"region1"`
	Region2 RegionAddressSize `json:"region2"`
	Region3 RegionAddressSize `json:"region3"`
	Region4 RegionAddressSize `json:"region4"`
	Region5 RegionAddressSize `json:"region5"`
	Region6 RegionAddressSize `json:"region6"`
}

type ResourceEntry struct {
	Offset   uint32 `json:"offset"`
	Size     uint32 `json:"size"`
	Unknown0 uint32 `json:"unknown0"`
}

type MDBContainer struct {
	Total   uint32          `json:"total"`
	Entries []ResourceEntry `json:"entries" length:"Total"`
}

type T32Container = MDBContainer

type Container struct {
	Offset       uint32       `json:"offset" skip:""`
	Header       Header       `json:"header"`
	Regions      Regions      `json:"regions"`
	MDBContainer MDBContainer `json:"mdb_container"`
	T32Container T32Container `json:"t32_container"`
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
		return fmt.Errorf("PKK offset %#x is outside stream size %#x",
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

	if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, &c.Regions); err != nil {
		return err
	}

	if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, &c.MDBContainer); err != nil {
		return err
	}

	if _, err := stream.Seek(int64(600-(c.MDBContainer.Total*12)), io.SeekCurrent); err != nil {
		return err
	}

	if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, &c.T32Container); err != nil {
		return err
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
