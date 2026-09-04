package t32

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"runtime"
	"sync"
	"sync/atomic"
)

const provenanceTrailerSize = 16

var (
	provenanceMagic    = [8]byte{'T', '3', '2', 'S', 'R', 'C', 0, 1}
	provenanceSequence atomic.Uint64
	provenanceRegistry sync.Map
)

func (d *DecodedTexture) PalettedImage() *image.Paletted {
	if d == nil {
		return nil
	}

	palette := make(color.Palette, len(d.Palette))
	for index, entry := range d.Palette {
		palette[index] = color.NRGBA{
			R: entry.R,
			G: entry.G,
			B: entry.B,
			A: entry.A,
		}
	}
	pixelTotal := d.Width * d.Height
	result := &image.Paletted{
		Pix:     make([]byte, pixelTotal, pixelTotal+provenanceTrailerSize),
		Stride:  d.Width,
		Rect:    image.Rect(0, 0, d.Width, d.Height),
		Palette: palette,
	}
	copy(result.Pix, d.Indices)
	if d.provenance != nil && fingerprintDecoded(d) == d.provenance.fingerprint {
		attachProvenance(result, d.provenance)
	}
	return result
}

func (c *Container) AddTexture(img *image.Paletted) (int, error) {
	provenance := provenanceFromPaletted(img)
	format := inferPixelStorageMode(img)
	if provenance != nil {
		format = provenance.source.textures[provenance.textureIndex].GsTex0.PSM
	}
	decoded, err := decodedFromPaletted(img, format)
	if err != nil {
		return 0, err
	}

	index := len(c.Textures)
	compatible := provenance != nil && provenance.textureIndex == index &&
		provenance.paletteIndex == 0
	if index == 0 && compatible && c.source == nil {
		c.source = provenance.source
		c.sourceExact = true
		c.Header = provenance.source.header
	}
	compatible = compatible && c.sourceExact && c.source == provenance.source
	if !compatible {
		c.sourceExact = false
	}

	texture := Texture{}
	if compatible {
		texture = c.source.textures[index]
	} else {
		texture.GsTex0.PSM = format
		texture.GsTex0Raw = uint64(format) << 20
		texture.GsTex1Raw = 0x260
		texture.GsTex1 = DecodeGSTex1(texture.GsTex1Raw)
	}
	decoded.TextureIndex = index
	decoded.PaletteIndex = 0
	decoded.provenance = provenance
	texture.Palettes = []DecodedTexture{*decoded}
	texture.PaletteTotal = 1
	texture.Index = uint32(index)
	c.Textures = append(c.Textures, texture)
	c.normalizeHierarchy()
	return index, nil
}

func (c *Container) AddPaletteAtTexture(index int, img *image.Paletted) error {
	texture, err := c.textureAt(index)
	if err != nil {
		return err
	}
	if len(texture.Palettes) == 0 {
		return fmt.Errorf("texture %d has no base palette", index)
	}
	base := &texture.Palettes[0]
	if img == nil || img.Bounds().Dx() != base.Width || img.Bounds().Dy() != base.Height {
		return fmt.Errorf(
			"texture %d palette dimensions must be %dx%d",
			index,
			base.Width,
			base.Height,
		)
	}

	provenance := provenanceFromPaletted(img)
	paletteIndex := len(texture.Palettes)
	compatible := provenance != nil && c.sourceExact && c.source == provenance.source &&
		provenance.textureIndex == index && provenance.paletteIndex == paletteIndex
	if !compatible {
		c.sourceExact = false
	}
	decoded, err := decodedFromPaletted(img, base.Format)
	if err != nil {
		return err
	}
	decoded.TextureIndex = index
	decoded.PaletteIndex = paletteIndex
	decoded.Indices = append(decoded.Indices[:0], base.Indices...)
	decoded.Pixels = pixelsFromIndices(decoded.Indices, decoded.Palette)
	decoded.provenance = provenance
	texture.Palettes = append(texture.Palettes, *decoded)
	c.normalizeHierarchy()

	return nil
}

func (c *Container) RemoveTexture(index int) error {
	if _, err := c.textureAt(index); err != nil {
		return err
	}
	copy(c.Textures[index:], c.Textures[index+1:])
	c.Textures[len(c.Textures)-1] = Texture{}
	c.Textures = c.Textures[:len(c.Textures)-1]
	c.sourceExact = false
	c.normalizeHierarchy()

	return nil
}

func (c *Container) RemovePaletteAtTexture(indexTexture, indexPalette int) error {
	texture, err := c.textureAt(indexTexture)
	if err != nil {
		return err
	}
	if indexPalette < 0 || indexPalette >= len(texture.Palettes) {
		return fmt.Errorf(
			"texture %d palette %d outside [0,%d)",
			indexTexture,
			indexPalette,
			len(texture.Palettes),
		)
	}
	copy(texture.Palettes[indexPalette:], texture.Palettes[indexPalette+1:])
	texture.Palettes[len(texture.Palettes)-1] = DecodedTexture{}
	texture.Palettes = texture.Palettes[:len(texture.Palettes)-1]
	c.sourceExact = false
	c.normalizeHierarchy()

	return nil
}

func (c *Container) normalizeHierarchy() {
	c.TextureTotal = uint32(len(c.Textures))
	paletteTotal := 0
	for textureIndex := range c.Textures {
		texture := &c.Textures[textureIndex]
		texture.Index = uint32(textureIndex)
		texture.PaletteTotal = uint32(len(texture.Palettes))
		paletteTotal += len(texture.Palettes)
		for paletteIndex := range texture.Palettes {
			decoded := &texture.Palettes[paletteIndex]
			decoded.TextureIndex = textureIndex
			decoded.PaletteIndex = paletteIndex
		}
	}
	c.PalettePacketTotal = uint16(paletteTotal)
}

func provenanceFromPaletted(img *image.Paletted) *decodedProvenance {
	if img == nil || len(img.Palette) == 0 ||
		cap(img.Pix)-len(img.Pix) < provenanceTrailerSize {
		return nil
	}
	extended := img.Pix[:len(img.Pix)+provenanceTrailerSize]
	trailer := extended[len(img.Pix):]
	if !bytes.Equal(trailer[:len(provenanceMagic)], provenanceMagic[:]) {
		return nil
	}
	id := binary.LittleEndian.Uint64(trailer[len(provenanceMagic):])
	value, ok := provenanceRegistry.Load(id)
	if !ok {
		return nil
	}
	provenance, ok := value.(*decodedProvenance)
	if !ok || provenance == nil || provenance.source == nil {
		return nil
	}
	if provenance.textureIndex < 0 ||
		provenance.textureIndex >= len(provenance.source.textures) {
		return nil
	}
	decoded, err := decodedFromPaletted(
		img,
		provenance.source.textures[provenance.textureIndex].GsTex0.PSM,
	)
	if err != nil || fingerprintDecoded(decoded) != provenance.fingerprint {
		return nil
	}
	return provenance
}

func attachProvenance(img *image.Paletted, provenance *decodedProvenance) {
	id := provenanceSequence.Add(1)
	provenanceRegistry.Store(id, provenance)
	extended := img.Pix[:len(img.Pix)+provenanceTrailerSize]
	trailer := extended[len(img.Pix):]
	copy(trailer, provenanceMagic[:])
	binary.LittleEndian.PutUint64(trailer[len(provenanceMagic):], id)
	runtime.AddCleanup(img, cleanupProvenance, id)
}

func cleanupProvenance(id uint64) {
	provenanceRegistry.Delete(id)
}

func inferPixelStorageMode(img *image.Paletted) PixelStorageMode {
	if img != nil && len(img.Palette) <= 16 {
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			offset := img.PixOffset(bounds.Min.X, y)
			for _, value := range img.Pix[offset : offset+bounds.Dx()] {
				if value > 0x0f {
					return PSMT8
				}
			}
		}
		return PSMT4
	}
	return PSMT8
}

func decodedFromPaletted(img *image.Paletted, format PixelStorageMode) (*DecodedTexture, error) {
	if img == nil {
		return nil, errors.New("paletted image is nil")
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("paletted image has invalid dimensions %dx%d", width, height)
	}
	paletteLimit := 256
	if format == PSMT4 {
		paletteLimit = 16
	} else if format != PSMT8 {
		return nil, fmt.Errorf("unsupported indexed format %s", format)
	}
	if len(img.Palette) == 0 || len(img.Palette) > paletteLimit {
		return nil, fmt.Errorf(
			"%s image must have between 1 and %d palette entries, got %d",
			format,
			paletteLimit,
			len(img.Palette),
		)
	}

	indices := make([]byte, width*height)
	for y := 0; y < height; y++ {
		offset := img.PixOffset(bounds.Min.X, bounds.Min.Y+y)
		copy(indices[y*width:(y+1)*width], img.Pix[offset:offset+width])
	}
	for pixel, value := range indices {
		if int(value) >= len(img.Palette) {
			return nil, fmt.Errorf(
				"pixel %d uses palette index %d outside %d entries",
				pixel,
				value,
				len(img.Palette),
			)
		}
	}

	palette := make([]RGBA, len(img.Palette))
	for index, entry := range img.Palette {
		converted := color.NRGBAModel.Convert(entry).(color.NRGBA)
		palette[index] = RGBA{
			R: converted.R,
			G: converted.G,
			B: converted.B,
			A: converted.A,
		}
	}
	return &DecodedTexture{
		Width:   width,
		Height:  height,
		Format:  format,
		Indices: indices,
		Palette: palette,
		Pixels:  pixelsFromIndices(indices, palette),
	}, nil
}

func pixelsFromIndices(indices []byte, palette []RGBA) []RGBA {
	pixels := make([]RGBA, len(indices))
	for index, value := range indices {
		if int(value) < len(palette) {
			pixels[index] = palette[value]
		}
	}
	return pixels
}
