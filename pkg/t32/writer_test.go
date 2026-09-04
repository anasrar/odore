package t32_test

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/anasrar/odore/pkg/t32"
)

func TestWriteAndDecodeCanonicalContainer(t *testing.T) {
	first := testPaletted(128, 128, 16, 1)
	secondPalette := testPaletted(128, 128, 16, 9)
	copy(secondPalette.Pix, first.Pix)
	second := testPaletted(256, 128, 256, 3)

	container := t32.New()
	if index, _ := container.AddTexture(first); index != 0 {
		t.Fatalf("first texture index = %d, want 0", index)
	}
	container.AddPaletteAtTexture(0, secondPalette)
	if index, _ := container.AddTexture(second); index != 1 {
		t.Fatalf("second texture index = %d, want 1", index)
	}

	var encoded bytes.Buffer
	if err := container.Write(&encoded); err != nil {
		t.Fatal(err)
	}

	decoded := t32.New()
	if err := t32.FromStream(decoded, bytes.NewReader(encoded.Bytes())); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Textures) != 2 {
		t.Fatalf("texture total = %d, want 2", len(decoded.Textures))
	}
	if len(decoded.Textures[0].Palettes) != 2 {
		t.Fatalf("texture 0 palette total = %d, want 2", len(decoded.Textures[0].Palettes))
	}
	assertPalettedEqual(t, decoded.Textures[0].Palettes[0].PalettedImage(), first)
	assertPalettedEqual(t, decoded.Textures[0].Palettes[1].PalettedImage(), secondPalette)
	assertPalettedEqual(t, decoded.Textures[1].Palettes[0].PalettedImage(), second)
}

func TestDecodedPalettedImagesRebuildIdenticalBytes(t *testing.T) {
	originalContainer := t32.New()
	originalContainer.AddTexture(testPaletted(128, 128, 16, 2))
	originalContainer.AddPaletteAtTexture(0, testPaletted(128, 128, 16, 8))
	originalContainer.AddTexture(testPaletted(256, 256, 256, 4))

	var original bytes.Buffer
	if err := originalContainer.Write(&original); err != nil {
		t.Fatal(err)
	}
	decoded := t32.New()
	if err := t32.FromStream(decoded, bytes.NewReader(original.Bytes())); err != nil {
		t.Fatal(err)
	}

	rebuilt := t32.New()
	for textureIndex := range decoded.Textures {
		for paletteIndex := range decoded.Textures[textureIndex].Palettes {
			img := decoded.Textures[textureIndex].Palettes[paletteIndex].PalettedImage()
			if paletteIndex == 0 {
				rebuilt.AddTexture(img)
			} else {
				rebuilt.AddPaletteAtTexture(textureIndex, img)
			}
		}
	}
	var output bytes.Buffer
	if err := rebuilt.Write(&output); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output.Bytes(), original.Bytes()) {
		t.Fatal("decode/add/write bytes differ")
	}
}

func TestRemoveTextureAndPalette(t *testing.T) {
	container := t32.New()
	container.AddTexture(testPaletted(128, 128, 16, 1))
	container.AddPaletteAtTexture(0, testPaletted(128, 128, 16, 2))
	container.AddTexture(testPaletted(128, 128, 256, 3))

	container.RemovePaletteAtTexture(0, 1)
	container.RemoveTexture(1)
	if len(container.Textures) != 1 || len(container.Textures[0].Palettes) != 1 {
		t.Fatalf("unexpected hierarchy after removal: %+v", container.Textures)
	}

	var encoded bytes.Buffer
	if err := container.Write(&encoded); err != nil {
		t.Fatal(err)
	}
	decoded := t32.New()
	if err := t32.FromStream(decoded, bytes.NewReader(encoded.Bytes())); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Textures) != 1 || len(decoded.Textures[0].Palettes) != 1 {
		t.Fatalf("unexpected decoded hierarchy: %+v", decoded.Textures)
	}
}

func TestCanonicalWriterSplitsLargeImageDMA(t *testing.T) {
	container := t32.New()
	images := make([]*image.Paletted, 8)
	for index := range images {
		images[index] = testPaletted(256, 256, 256, index+1)
		container.AddTexture(images[index])
	}

	var encoded bytes.Buffer
	if err := container.Write(&encoded); err != nil {
		t.Fatal(err)
	}
	decoded := t32.New()
	if err := t32.FromStream(decoded, bytes.NewReader(encoded.Bytes())); err != nil {
		t.Fatal(err)
	}
	if len(decoded.ImageDMA.Transfers) != 2 {
		t.Fatalf("image DMA transfer total = %d, want 2", len(decoded.ImageDMA.Transfers))
	}
	assertPalettedEqual(t, decoded.Textures[0].Palettes[0].PalettedImage(), images[0])
	assertPalettedEqual(t, decoded.Textures[7].Palettes[0].PalettedImage(), images[7])
}

func testPaletted(width, height, paletteSize, seed int) *image.Paletted {
	palette := make(color.Palette, paletteSize)
	for index := range palette {
		palette[index] = color.NRGBA{
			R: uint8(index*3 + seed),
			G: uint8(index*5 + seed*2),
			B: uint8(index*7 + seed*3),
			A: uint8((index*2 + seed*4) & 0xfe),
		}
	}
	palette[paletteSize-1] = color.NRGBA{R: 0xff, G: 0xfe, B: 0xfd, A: 0xff}
	result := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	for index := range result.Pix {
		result.Pix[index] = uint8((index*7 + seed) % paletteSize)
	}
	return result
}

func assertPalettedEqual(t *testing.T, actual, expected *image.Paletted) {
	t.Helper()
	if actual.Bounds() != expected.Bounds() {
		t.Fatalf("bounds = %v, want %v", actual.Bounds(), expected.Bounds())
	}
	if !bytes.Equal(actual.Pix, expected.Pix) {
		t.Fatal("palette indices differ")
	}
	if len(actual.Palette) != len(expected.Palette) {
		t.Fatalf("palette size = %d, want %d", len(actual.Palette), len(expected.Palette))
	}
	for index := range actual.Palette {
		got := color.NRGBAModel.Convert(actual.Palette[index]).(color.NRGBA)
		want := color.NRGBAModel.Convert(expected.Palette[index]).(color.NRGBA)
		if got != want {
			t.Fatalf("palette %d = %+v, want %+v", index, got, want)
		}
	}
}
