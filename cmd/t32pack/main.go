package main

import (
	"encoding/json"
	"image"
	"image/png"
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
	Use:   "t32pack",
	Short: "Pack PNG to T32",
	Run: func(cmd *cobra.Command, args []string) {
		metadataBuf, err := os.ReadFile(input)
		if err != nil {
			log.Fatal(err)
		}

		container := t32.New()

		m := metadata.MetadataT32{}
		if err := json.Unmarshal(metadataBuf, &m); err != nil {
			log.Fatal(err)
		}

		for textureIndex, texture := range m.Textures {
			ti := textureIndex
			for paletteIndex, palette := range texture.Palettes {
				pngFile, err := os.Open(palette.Path)
				if err != nil {
					log.Fatal(err)
				}
				defer pngFile.Close()

				img, err := png.Decode(pngFile)
				if err != nil {
					log.Fatal(err)
				}

				imgPaletted, ok := img.(*image.Paletted)
				if !ok {
					log.Fatal("not indexed PNG")
				}

				if paletteIndex == 0 {
					ti = container.AddTexture(imgPaletted)
				} else {
					container.AddPaletteAtTexture(ti, imgPaletted)
				}

				log.Printf("inserted: texture %d palette %d", texture.Index, palette.Index)
			}
		}

		outputPath := filepath.Join(utils.PackPath(input), "OUTPUT.t32")

		file, err := os.Create(outputPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		if err := container.Write(file); err != nil {
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
