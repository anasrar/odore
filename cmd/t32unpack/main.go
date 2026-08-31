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
	"github.com/spf13/cobra"
)

var input string

var cmd = &cobra.Command{
	Use:   "t32unpack",
	Short: "Unpack T32 to PNG",
	Run: func(cmd *cobra.Command, args []string) {
		pathDirectoryFiles, pathManifest, err := utils.UnpackPath(input)
		if err != nil {
			log.Fatal(err)
		}

		if err := os.MkdirAll(pathDirectoryFiles, 0755); err != nil {
			log.Fatal(err)
		}

		file, err := os.Open(input)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		container := t32.New()
		if err := t32.FromStream(container, file); err != nil {
			log.Fatal(err)
		}

		m := metadata.MetadataT32{Textures: make([]metadata.TextureT32, 0)}

		for textureIndex, texture := range container.Textures {
			if !texture.IsIndexed() {
				log.Printf(
					"skip texture %d: pixel format %#x is not indexed",
					textureIndex,
					uint8(texture.GsTex0.PSM),
				)
				continue
			}

			tx := metadata.TextureT32{
				Index:    uint64(texture.Index),
				Palettes: make([]metadata.PaletteT32, 0),
			}

			for paletteIndex := range texture.Palettes {
				decoded, err := container.DecodeTexture(textureIndex, paletteIndex)
				if err != nil {
					log.Fatal(err)
				}

				filename := fmt.Sprintf("texture_%03d_palette_%03d.png", textureIndex, paletteIndex)
				outputPath := filepath.Join(pathDirectoryFiles, filename)
				if err := utils.WritePNG(outputPath, decoded.PalettedImage()); err != nil {
					log.Fatal(err)
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
			log.Fatal(err)
		}
		defer file.Close()

		j, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		if _, err := fileManifest.Write(j); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	cmd.Flags().StringVarP(&input, "input", "i", "", "Input file")
	cmd.MarkFlagRequired("input")
}

func main() {
	cmd.Execute()
}
