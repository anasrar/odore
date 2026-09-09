package mdb

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/anasrar/binarium"
	"github.com/anasrar/odore/pkg/utils"
	"github.com/qmuntal/gltf"
)

type Header struct {
	Signature              uint32   `json:"signature"`
	BoneOffset             uint32   `json:"bone_offset"`
	BoneTotal              uint16   `json:"bone_total"`
	VertexBufferTotal      uint16   `json:"vertex_buffer_total"`
	Unknown0               uint32   `json:"unknown0"`
	SkinningFormat         uint32   `json:"skinning_format"`
	BoneIndexTableOffset   uint32   `json:""`
	BonePaletteTableOffset uint32   `json:"bone_palette_table_offset"`
	Scale                  float32  `json:"scale"`
	VertexBufferOffsets    []uint32 `json:"vertex_buffer_offsets" length:"VertexBufferTotal"`
}

type Container struct {
	Offset        uint32                   `json:"offset" skip:""`
	Header        Header                   `json:"header"`
	VertexBuffers []*VertexBufferContainer `json:"vertex_buffers"`
}

func New() *Container {
	return &Container{
		VertexBuffers: []*VertexBufferContainer{},
	}
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
		return fmt.Errorf("MDB offset %#x is outside stream size %#x",
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

	for _, offset := range c.Header.VertexBufferOffsets {
		vbOffset := uint64(offset)

		if err := utils.SeekAbsolute(stream, baseOffset+vbOffset); err != nil {
			return err
		}

		vbContainer := NewVertexBufferContainer()
		if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, &vbContainer.Header); err != nil {
			return err
		}

		vbContainer.ContainerPositions.Total = vbContainer.Header.VertexTotal
		if err := utils.SeekAbsolute(stream, baseOffset+vbOffset+uint64(vbContainer.Header.PositionOffset)); err != nil {
			return err
		}
		if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, &vbContainer.ContainerPositions); err != nil {
			return err
		}

		c.VertexBuffers = append(c.VertexBuffers, vbContainer)
	}

	return nil
}

func (c *Container) ConvrtToGLTF(doc *gltf.Document, mdbIndex int) error {
	name := fmt.Sprintf("mdb_%03d", mdbIndex)

	mesh := &gltf.Mesh{
		Name:       name,
		Primitives: []*gltf.Primitive{},
	}

	for _, entry := range c.VertexBuffers {
		primitive := entry.ConvertToGLTFPrimitive(doc, c.Header.Scale)
		mesh.Primitives = append(mesh.Primitives, primitive)
	}

	doc.Meshes = append(doc.Meshes, mesh)

	meshIndex := gltf.Index(len(doc.Meshes) - 1)
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name: name,
		Mesh: meshIndex,
	})

	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, len(doc.Nodes)-1)

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
