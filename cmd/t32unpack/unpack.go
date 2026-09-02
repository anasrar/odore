package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/anasrar/odore/pkg/metadata"
	"github.com/anasrar/odore/pkg/t32"
	"github.com/anasrar/odore/pkg/utils"
)

func unpack(input string, offset uint32) error {
	pathDirectoryFiles, pathManifest, err := utils.UnpackPath(input)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(pathDirectoryFiles, 0755); err != nil {
		return (err)
	}

	container := t32.New()
	if err := t32.FromPathWithOffset(container, input, offset); err != nil {
		return (err)
	}

	m := metadata.MetadataT32{Textures: make([]metadata.TextureT32, 0)}

	for textureIndex, texture := range container.Textures {
		if !texture.IsIndexed() {
			return fmt.Errorf(
				"skip texture %d: pixel format %#x is not indexed",
				textureIndex,
				uint8(texture.GsTex0.PSM),
			)
		}

		tx := metadata.TextureT32{
			Index:    uint64(texture.Index),
			Palettes: make([]metadata.PaletteT32, 0),
		}

		for paletteIndex := range texture.Palettes {
			decoded, err := container.DecodeTexture(textureIndex, paletteIndex)
			if err != nil {
				return (err)
			}

			filename := fmt.Sprintf("texture_%03d_palette_%03d.png", textureIndex, paletteIndex)
			outputPath := filepath.Join(pathDirectoryFiles, filename)
			if err := utils.WritePNG(outputPath, decoded.PalettedImage()); err != nil {
				return (err)
			}

			log.Printf("create: %s", filename)

			tx.Palettes = append(tx.Palettes, metadata.PaletteT32{
				Index: uint64(paletteIndex),
				Path:  outputPath,
			})
		}

		m.Textures = append(m.Textures, tx)
	}

	fileManifest, err := os.Create(pathManifest)
	if err != nil {
		return (err)
	}
	defer fileManifest.Close()

	j, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		return (err)
	}

	if _, err := fileManifest.Write(j); err != nil {
		return (err)
	}

	return nil
}
