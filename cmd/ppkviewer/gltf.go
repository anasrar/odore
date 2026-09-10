package main

import (
	"bytes"
	"fmt"

	"github.com/anasrar/odore/pkg/utils"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

func convrtToGLTF() error {
	doc := gltf.NewDocument()
	zero := float64(0)
	one := float64(1)

	for t32Index, entry := range textures {
		texIndex, err := modeler.WriteImage(doc, fmt.Sprintf("%03d", t32Index), "image/png", bytes.NewReader(entry.PNG))
		if err != nil {
			return err
		}
		doc.Textures = append(doc.Textures, &gltf.Texture{
			Source: gltf.Index(texIndex),
		})

		doc.Materials = append(doc.Materials,
			&gltf.Material{
				PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
					BaseColorTexture: &gltf.TextureInfo{
						Index: texIndex,
					},
					MetallicFactor:  &zero,
					RoughnessFactor: &one,
				},
				AlphaMode: gltf.AlphaMask,
			},
		)
	}

	doc.Meshes = []*gltf.Mesh{}

	for mdbIndex, entry := range mdbContainers {
		if err := entry.ConvrtToGLTF(doc, mdbIndex); err != nil {
			return err
		}
	}

	gltf.SaveBinary(doc, utils.GLTFPath(Input))

	return nil
}
