package mdb

import (
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

type VertexBufferHeader struct {
	PositionOffset uint32 `json:"position_offset"`
	NormalOffset   uint32 `json:"normal_offset"`
	UVOffset       uint32 `json:"uv_offset"`
	ColorOffset    uint32 `json:"color_offset"`
	WeightOffset   uint32 `json:"weight_offset"`
	VertexTotal    uint16 `json:"vertex_total"`
	Unknown0       uint8  `json:"unknown0"`
	Material       uint8  `json:"material"`
}

type VertexBufferContainer struct {
	Header             VertexBufferHeader `json:"header"`
	ContainerPositions PositionContainer  `json:"positions"`
	ContainerNormals   NormalContainer    `json:"normals"`
}

func (vb *VertexBufferContainer) ConvertToGLTFPrimitive(doc *gltf.Document, scale float32) *gltf.Primitive {
	vertices := [][3]float32{}
	indecies := []uint16{}

	for i, pos := range vb.ContainerPositions.Entries {
		ii := uint16(i)
		vertices = append(vertices, [3]float32{pos.X * scale, pos.Y * scale, pos.Z * scale})

		if pos.Flag == 0x8000 {
			continue
		} else if pos.Flag == 0x0 {
			indecies = append(indecies, ii-2, ii-1, ii)
		} else if pos.Flag == 0x1 {
			indecies = append(indecies, ii-1, ii-2, ii)
		}
	}

	return &gltf.Primitive{
		Indices: gltf.Index(modeler.WriteIndices(doc, indecies)),
		Attributes: gltf.PrimitiveAttributes{
			gltf.POSITION: modeler.WritePosition(doc, vertices),
		},
	}
}

func NewVertexBufferContainer() *VertexBufferContainer {
	return &VertexBufferContainer{}
}
