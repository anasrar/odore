package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/anasrar/odore/pkg/manifest"
	"github.com/anasrar/odore/pkg/t32"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "t32topng [T32]",
		Short: "Convert T32 to PNG",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			pathFromArgs := args[0]
			pathDirectoryFiles, pathManifest, err := unpackPath(pathFromArgs)
			if err != nil {
				log.Fatal(err)
			}

			if err := os.MkdirAll(pathDirectoryFiles, 0755); err != nil {
				log.Fatal(err)
			}

			file, err := os.Open(pathFromArgs)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			container := t32.New()
			if err := t32.FromStream(container, file); err != nil {
				log.Fatal(err)
			}

			metadata := manifest.Metadata{Textures: make([]manifest.Texture, 0)}

			for textureIndex, texture := range container.Textures {
				if !texture.IsIndexed() {
					log.Printf(
						"skip texture %d: pixel format %#x is not indexed",
						textureIndex,
						uint8(texture.GsTex0.PSM),
					)
					continue
				}

				tx := manifest.Texture{
					Index:    uint64(texture.Index),
					Palettes: make([]manifest.Palette, 0),
				}

				for paletteIndex := 0; paletteIndex < int(texture.PaletteTotal); paletteIndex++ {
					decoded, err := container.DecodeTexture(textureIndex, paletteIndex)
					if err != nil {
						log.Fatal(err)
					}

					filename := fmt.Sprintf("texture_%03d_palette_%03d.png", textureIndex, paletteIndex)
					outputPath := filepath.Join(pathDirectoryFiles, filename)
					if err := writePNG(outputPath, decoded.Image()); err != nil {
						log.Fatal(err)
					}

					log.Printf("create: %s", filename)

					tx.Palettes = append(tx.Palettes, manifest.Palette{
						Index: uint64(paletteIndex),
						Path:  outputPath,
					})
				}

				metadata.Textures = append(metadata.Textures, tx)
			}

			fileManifest, err := os.Create(pathManifest)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			j, err := json.MarshalIndent(metadata, "", "  ")
			if err != nil {
				log.Fatal(err)
			}

			if _, err := fileManifest.Write(j); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Execute()
}

func writePNG(outputPath string, imageData image.Image) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := png.Encode(file, imageData); err != nil {
		return err
	}
	return nil
}

func unpackPath(path string) (string, string, error) {
	filename := fmt.Sprintf("UNPACK_%s", filepath.Base(path))
	directory := ""

	if filepath.IsAbs(path) {
		directory = filepath.Dir(filepath.Clean(path))
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return "", "", err
		}

		directory = filepath.Dir(filepath.Clean(filepath.Join(cwd, path)))
	}

	pathDirectoryFiles := filepath.Join(directory, filename, "FILES")
	pathManifest := filepath.Join(directory, filename, "MANIFEST.json")

	return pathDirectoryFiles, pathManifest, nil
}
