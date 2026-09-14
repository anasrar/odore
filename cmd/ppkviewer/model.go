package main

import (
	"unsafe"

	"github.com/anasrar/odore/pkg/mdb"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Model struct {
	Meshes       []rl.Mesh
	TextureIndex int32
	Material     rl.Material
}

func ModelNew(container *mdb.Container) *Model {
	model := &Model{
		Meshes:       []rl.Mesh{},
		TextureIndex: 0,
		Material:     rl.LoadMaterialDefault(),
	}

	for _, vb := range container.VertexBuffers {
		vertices := []float32{}
		indices := []uint16{}

		for i, position := range vb.ContainerPositions.Entries {
			vertices = append(vertices,
				position.X*container.Header.Scale,
				position.Y*container.Header.Scale,
				position.Z*container.Header.Scale,
			)

			ii := uint16(i)
			if i < 2 || position.Flag == 0x8000 {
				continue
			} else if position.Flag == 0x0 {
				indices = append(indices, ii-2, ii-1, ii)
			} else if position.Flag == 0x1 {
				indices = append(indices, ii-1, ii-2, ii)
			}
		}

		if len(indices) == 0 {
			continue
		}

		mesh := rl.Mesh{
			VertexCount:   int32(len(vertices) / 3),
			TriangleCount: int32(len(indices) / 3),
			Vertices:      unsafe.SliceData(vertices),
			Indices:       unsafe.SliceData(indices),
		}

		if vb.Header.NormalOffset != 0 {
			normals := []float32{}
			for _, normal := range vb.ContainerNormals.Entries {
				normals = append(normals, normal.X, normal.Y, normal.Z)
			}
			mesh.Normals = unsafe.SliceData(normals)
		}

		if vb.Header.UVOffset != 0 {
			uvs := []float32{}
			for _, uv := range vb.ContainerUVs.Entries {
				uvs = append(uvs, uv.U, uv.V+1)
			}
			mesh.Texcoords = unsafe.SliceData(uvs)
		}

		rl.UploadMesh(&mesh, false)
		model.Meshes = append(model.Meshes, mesh)
	}

	return model
}

func (m *Model) Draw() {
	texture := rl.Texture2D{}
	if m.TextureIndex >= 0 && m.TextureIndex < int32(len(textures)) {
		texture = textures[m.TextureIndex].Texture
	}

	materialMap := m.Material.GetMap(rl.MapDiffuse)
	previousTexture := materialMap.Texture
	if texture.ID != 0 {
		materialMap.Texture = texture
	}

	for _, mesh := range m.Meshes {
		rl.DrawMesh(mesh, m.Material, rl.MatrixIdentity())
	}

	materialMap.Texture = previousTexture
}

func (m *Model) Unload() {
	for i := range m.Meshes {
		rl.UnloadMesh(&m.Meshes[i])
	}
	rl.UnloadMaterial(m.Material)
}
