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
	ContainerUVs       UVContainer        `json:"uvs"`
}

func (vb *VertexBufferContainer) ConvertToGLTFPrimitive(doc *gltf.Document, scale float32) *gltf.Primitive {
	vertices := [][3]float32{}
	indices := []uint16{}

	for i, pos := range vb.ContainerPositions.Entries {
		ii := uint16(i)
		vertices = append(vertices, [3]float32{pos.X * scale, pos.Y * scale, pos.Z * scale})

		if pos.Flag == 0x8000 {
			continue
		} else if pos.Flag == 0x0 {
			indices = append(indices, ii-2, ii-1, ii)
		} else if pos.Flag == 0x1 {
			indices = append(indices, ii-1, ii-2, ii)
		}
	}

	attributes := gltf.PrimitiveAttributes{
		gltf.POSITION: modeler.WritePosition(doc, vertices),
	}

	if vb.Header.NormalOffset != 0 {
		normals := make([][3]float32, vb.ContainerNormals.Total)
		for i, normal := range vb.ContainerNormals.Entries {
			normals[i] = [3]float32{normal.X, normal.Y, normal.Z}
		}
		attributes[gltf.NORMAL] = modeler.WriteNormal(doc, normals)
	}

	return &gltf.Primitive{
		Indices:    gltf.Index(modeler.WriteIndices(doc, indices)),
		Attributes: attributes,
	}
}

func NewVertexBufferContainer() *VertexBufferContainer {
	return &VertexBufferContainer{}
}
