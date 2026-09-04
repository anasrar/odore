package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/anasrar/odore/pkg/metadata"
	"github.com/anasrar/odore/pkg/t32"
	"github.com/anasrar/odore/pkg/utils"
)

func pack(input string) error {
	metadataBuf, err := os.ReadFile(input)
	if err != nil {
		return err
	}

	container := t32.New()

	m := metadata.MetadataT32{}
	if err := json.Unmarshal(metadataBuf, &m); err != nil {
		return err
	}

	for textureIndex, texture := range m.Textures {
		ti := textureIndex
		for paletteIndex, palette := range texture.Palettes {
			pngFile, err := os.Open(palette.Path)
			if err != nil {
				return err
			}
			defer pngFile.Close()

			img, err := png.Decode(pngFile)
			if err != nil {
				return err
			}

			imgPaletted, ok := img.(*image.Paletted)
			if !ok {
				return fmt.Errorf("not indexed PNG")
			}

			if paletteIndex == 0 {
				ti, err = container.AddTexture(imgPaletted)
				if err != nil {
					return err
				}
			} else {
				container.AddPaletteAtTexture(ti, imgPaletted)
			}

			log.Printf("inserted: texture %d palette %d", texture.Index, palette.Index)
		}
	}

	outputPath := filepath.Join(utils.PackPath(input), "OUTPUT.t32")

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := container.Write(file); err != nil {
		return err
	}

	return nil
}
