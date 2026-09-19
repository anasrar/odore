package scd

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/anasrar/binarium"
	"github.com/anasrar/odore/pkg/utils"
)

type Header struct {
	Signature        uint32        `json:"signature"`
	Unknown0         uint32        `json:"unknown0"`
	Unknown1         uint32        `json:"unknown1"`
	SceneObjectTotal uint32        `json:"scene_object_total"`
	Unknown2         uint32        `json:"unknown2"`
	Unknown3         uint32        `json:"unknown3"`
	Unknown4         uint32        `json:"unknown4"`
	Unknown5         uint32        `json:"unknown5"`
	SceneObjects     []SceneObject `json:"scene_objects" length:"SceneObjectTotal"`
}

type Container struct {
	Offset uint32 `json:"offset" skip:""`
	Header Header `json:"header"`
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
		return fmt.Errorf("scd offset %#x is outside stream size %#x",
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
