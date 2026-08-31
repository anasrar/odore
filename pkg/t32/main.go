package t32

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"
	"os"

	"github.com/anasrar/binarium"
	"github.com/anasrar/graphicsynthesizer"
	"github.com/anasrar/odore/pkg/utils"
)

type Header struct {
	TextureTotal       uint32 `json:"texture_total"`
	TransferWidth      uint16 `json:"transfer_width"`
	TransferHeight     uint16 `json:"transfer_height"`
	ImageDMAOffset     uint32 `json:"image_dma_offset"`
	PaletteDMAOffset   uint32 `json:"palette_dma_offset"`
	ImagePacketTotal   uint16 `json:"image_packet_total"`
	PalettePacketTotal uint16 `json:"palette_packet_total"`
	Reserved0          uint32 `json:"reserved_0"`
	Reserved1          uint32 `json:"reserved_1"`
	FirstTextureSize   uint32 `json:"first_texture_size"`
}

type Texture struct {
	PaletteTotal     uint32 `json:"palette_total"`
	Reserved0        uint32 `json:"reserved_0"`
	GsTex0Raw        uint64 `json:"gs_tex_0_raw"`
	GsTex1Raw        uint32 `json:"gs_tex_1_raw"`
	Reserved1        uint32 `json:"reserved_1"`
	NextTextureIndex uint32 `json:"next_texture_index"`
	NextTextureSize  uint32 `json:"next_texture_size"`

	Index    uint32 `json:"index" skip:""`
	DataSize uint32 `json:"data_size" skip:""`
	GsTex0   GSTex0 `json:"gs_tex_0" skip:""`
	GsTex1   GSTex1 `json:"gs_tex_1" skip:""`
}

func (t *Texture) IsIndexed() bool {
	return t != nil && (t.GsTex0.PSM == PSMT8 || t.GsTex0.PSM == PSMT4)
}

func (t *Texture) PaletteEntryTotal() int {
	if t == nil {
		return 0
	}
	switch t.GsTex0.PSM {
	case PSMT8:
		return 256
	case PSMT4:
		return 16
	default:
		return 0
	}
}

type DecodedTexture struct {
	TextureIndex int              `json:"texture_index"`
	PaletteIndex int              `json:"palette_index"`
	Width        int              `json:"width"`
	Height       int              `json:"height"`
	Format       PixelStorageMode `json:"format"`
	Indices      []byte           `json:"indices"`
	Palette      []RGBA           `json:"palette"`
	Pixels       []RGBA           `json:"pixels"`
}

func (d *DecodedTexture) RGBABytes() []byte {
	if d == nil {
		return nil
	}
	data := make([]byte, len(d.Pixels)*4)
	for i, pixel := range d.Pixels {
		data[i*4+0] = pixel.R
		data[i*4+1] = pixel.G
		data[i*4+2] = pixel.B
		data[i*4+3] = pixel.A
	}
	return data
}

func (self *DecodedTexture) Image() *image.NRGBA {
	if self == nil {
		return nil
	}
	img := image.NewNRGBA(image.Rect(0, 0, self.Width, self.Height))
	copy(img.Pix, self.RGBABytes())
	return img
}

type Container struct {
	Offset     uint32 `json:"offset" skip:""`
	Header     `json:"header"`
	Textures   []Texture `json:"textures" length:"TextureTotal"`
	ImageDMA   DMAChain  `json:"image_dma" skip:""`
	PaletteDMA DMAChain  `json:"palette_dma" skip:""`
}

func New() *Container {
	return &Container{
		Textures:   []Texture{},
		ImageDMA:   DMAChain{Transfers: []DMATransfer{}},
		PaletteDMA: DMAChain{Transfers: []DMATransfer{}},
	}
}

func (c *Container) unmarshal(stream io.ReadSeeker) error {
	if c == nil {
		return ErrContainerIsNil
	}
	if stream == nil {
		return errors.New("stream is nil")
	}

	fileSize, err := utils.SeekerSize(stream)
	if err != nil {
		return err
	}
	baseOffset := uint64(c.Offset)
	if baseOffset > fileSize {
		return fmt.Errorf("T32 offset %#x is outside stream size %#x",
			baseOffset, fileSize)
	}

	if err := utils.SeekAbsolute(stream, baseOffset); err != nil {
		return err
	}
	var header Header
	if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("Header: %w", err)
	}
	if err := ValidateLayout(fileSize-baseOffset, header); err != nil {
		return err
	}

	parsed := New()
	parsed.Offset = c.Offset
	if err := utils.SeekAbsolute(stream, baseOffset); err != nil {
		return err
	}

	if err := binarium.UnmarshalWithReader(stream, binary.LittleEndian, parsed); err != nil {
		return fmt.Errorf("Container structure: %w", err)
	}

	if err := parsed.resolveTextures(); err != nil {
		return err
	}
	parsed.ImageDMA, err = ReadDMAChain(
		stream,
		fileSize,
		baseOffset,
		parsed.ImageDMAOffset,
		parsed.ImagePacketTotal,
		"image",
	)
	if err != nil {
		return err
	}
	parsed.PaletteDMA, err = ReadDMAChain(
		stream,
		fileSize,
		baseOffset,
		parsed.PaletteDMAOffset,
		parsed.PalettePacketTotal,
		"palette",
	)
	if err != nil {
		return err
	}

	*c = *parsed
	return nil
}

func (c *Container) DecodeTexture(textureIndex, paletteIndex int) (*DecodedTexture, error) {
	texture, err := c.textureAt(textureIndex)
	if err != nil {
		return nil, err
	}
	if !texture.IsIndexed() {
		return nil, ErrNotIndexed
	}

	indices, err := c.PixelIndices(textureIndex)
	if err != nil {
		return nil, err
	}
	palette, err := c.ColorPalette(textureIndex, paletteIndex)
	if err != nil {
		return nil, err
	}

	pixels := make([]RGBA, len(indices))
	for i, index := range indices {
		if int(index) >= len(palette) {
			return nil, fmt.Errorf(
				"texture %d pixel %d uses palette index %d outside %d entries",
				textureIndex,
				i,
				index,
				len(palette),
			)
		}
		pixels[i] = palette[index]
	}

	return &DecodedTexture{
		TextureIndex: textureIndex,
		PaletteIndex: paletteIndex,
		Width:        int(texture.GsTex0.Width()),
		Height:       int(texture.GsTex0.Height()),
		Format:       texture.GsTex0.PSM,
		Indices:      indices,
		Palette:      palette,
		Pixels:       pixels,
	}, nil
}

func (c *Container) PixelIndices(textureIndex int) ([]byte, error) {
	texture, err := c.textureAt(textureIndex)
	if err != nil {
		return nil, err
	}
	if !texture.IsIndexed() {
		return nil, ErrNotIndexed
	}

	gsMutex.Lock()
	defer gsMutex.Unlock()
	clear(graphicsynthesizer.GsMem)

	for i := range c.ImageDMA.Transfers {
		transfer := &c.ImageDMA.Transfers[i]
		upload := transfer.Upload
		if upload.Direction != TransferHostToLocal {
			return nil, fmt.Errorf("image DMA transfer %d is not host-to-local", i)
		}
		if upload.DestinationPSM != PSMCT32 {
			return nil, fmt.Errorf(
				"image DMA transfer %d has unsupported destination format %#x",
				i,
				uint8(upload.DestinationPSM),
			)
		}
		required := int(upload.Width) * int(upload.Height) * 4
		if len(transfer.Data) < required {
			return nil, fmt.Errorf(
				"image DMA transfer %d has %#x bytes, need %#x",
				i,
				len(transfer.Data),
				required,
			)
		}

		dbw := int(upload.DestinationBufferWidth)
		if dbw < 1 {
			dbw = 1
		}

		graphicsynthesizer.WriteTexPSMCT32(
			int(upload.DestinationBase),
			dbw,
			int(upload.DestinationX),
			int(upload.DestinationY),
			int(upload.Width),
			int(upload.Height),
			transfer.Data,
		)
	}

	width := int(texture.GsTex0.Width())
	height := int(texture.GsTex0.Height())
	pixelTotal := width * height
	tbw := int(texture.GsTex0.TBW)
	if tbw < 1 {
		tbw = 1
	}

	switch texture.GsTex0.PSM {
	case PSMT8:
		indices := make([]byte, pixelTotal)
		graphicsynthesizer.ReadTexPSMT8(
			int(texture.GsTex0.TBP0),
			tbw,
			0,
			0,
			width,
			height,
			indices,
		)
		return indices, nil

	case PSMT4:
		packed := make([]byte, (pixelTotal+1)/2)
		graphicsynthesizer.ReadTexPSMT4(
			int(texture.GsTex0.TBP0),
			tbw,
			0,
			0,
			width,
			height,
			packed,
		)
		indices := make([]byte, pixelTotal)
		for i := range indices {
			value := packed[i/2]
			if i&1 == 0 {
				indices[i] = value & 0x0f
			} else {
				indices[i] = value >> 4
			}
		}
		return indices, nil

	default:
		return nil, ErrNotIndexed
	}
}

func (c *Container) ColorPalette(textureIndex, paletteIndex int) ([]RGBA, error) {
	texture, err := c.textureAt(textureIndex)
	if err != nil {
		return nil, err
	}
	if !texture.IsIndexed() {
		return nil, ErrNotIndexed
	}
	if texture.GsTex0.CSM != 0 {
		return nil, fmt.Errorf("texture %d uses unsupported CSM2 palette storage", textureIndex)
	}
	if texture.GsTex0.CPSM != PSMCT32 {
		return nil, fmt.Errorf(
			"texture %d uses unsupported CLUT format %#x",
			textureIndex,
			uint8(texture.GsTex0.CPSM),
		)
	}
	if paletteIndex < 0 || uint32(paletteIndex) >= texture.PaletteTotal {
		return nil, fmt.Errorf(
			"texture %d palette %d outside [0,%d)",
			textureIndex,
			paletteIndex,
			texture.PaletteTotal,
		)
	}

	packetIndex := int(texture.GsTex0.CBP) + paletteIndex
	if packetIndex < 0 || packetIndex >= len(c.PaletteDMA.Transfers) {
		return nil, fmt.Errorf(
			"texture %d refers to palette DMA transfer %d outside %d transfers",
			textureIndex,
			packetIndex,
			len(c.PaletteDMA.Transfers),
		)
	}

	payload := c.PaletteDMA.Transfers[packetIndex].Data
	if len(payload) < 256*4 {
		return nil, fmt.Errorf(
			"palette DMA transfer %d has %#x bytes, need %#x",
			packetIndex,
			len(payload),
			256*4,
		)
	}

	fullPalette := DecodeCSM1RGBA32(payload)
	if texture.GsTex0.PSM == PSMT8 {
		return fullPalette, nil
	}

	base := int(texture.GsTex0.CSA&0x0f) * 16
	palette := make([]RGBA, 16)
	copy(palette, fullPalette[base:base+16])
	return palette, nil
}

func (c *Container) PixelColors(textureIndex, paletteIndex int) ([]RGBA, error) {
	decoded, err := c.DecodeTexture(textureIndex, paletteIndex)
	if err != nil {
		return nil, err
	}
	return decoded.Pixels, nil
}

func (c *Container) textureAt(index int) (*Texture, error) {
	if c == nil {
		return nil, errors.New("T32 is nil")
	}
	if index < 0 || index >= len(c.Textures) {
		return nil, fmt.Errorf("texture index %d outside [0,%d)", index, len(c.Textures))
	}
	return &c.Textures[index], nil
}

func (c *Container) resolveTextures() error {
	for i := range c.Textures {
		texture := &c.Textures[i]
		texture.Index = uint32(i)
		texture.GsTex0 = DecodeGSTex0(texture.GsTex0Raw)
		texture.GsTex1 = DecodeGSTex1(texture.GsTex1Raw)
	}

	if len(c.Textures) != 0 {
		c.Textures[0].DataSize = c.FirstTextureSize
	}
	for i := range c.Textures {
		texture := &c.Textures[i]
		if texture.NextTextureIndex == 0 && texture.NextTextureSize ==
			0 {
			continue
		}
		if texture.NextTextureIndex >= uint32(len(c.Textures)) {
			return fmt.Errorf(
				"next texture index %d is outside %d descriptors",
				texture.NextTextureIndex,
				len(c.Textures),
			)
		}
		c.Textures[texture.NextTextureIndex].DataSize = texture.NextTextureSize
	}
	return nil
}

func FromStreamWithOffset(c *Container, stream io.ReadSeeker, offset uint32) error {
	if c == nil {
		return ErrContainerIsNil
	}
	c.Offset = offset
	return c.unmarshal(stream)
}

func FromStream(c *Container, stream io.ReadSeeker) error {
	return FromStreamWithOffset(c, stream, 0)
}

func FromPathWithOffset(c *Container, filePath string, offset uint32) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return FromStreamWithOffset(c, file, offset)
}

func FromPath(c *Container, filePath string) error {
	return FromPathWithOffset(c, filePath, 0)
}

func ValidateLayout(dataSize uint64, header Header) error {
	descriptorSize := uint64(header.TextureTotal) * TextureEntrySize
	descriptorEnd := uint64(HeaderSize) + descriptorSize
	if descriptorEnd > dataSize {
		return fmt.Errorf(
			"texture descriptor table ends at %#x, T32 data size is %#x",
			descriptorEnd,
			dataSize,
		)
	}
	if uint64(header.ImageDMAOffset) < descriptorEnd {
		return fmt.Errorf(
			"image DMA offset %#x overlaps descriptor table ending at %#x",
			header.ImageDMAOffset,
			descriptorEnd,
		)
	}
	if uint64(header.ImageDMAOffset) >= dataSize {
		return fmt.Errorf(
			"image DMA offset %#x is outside T32 data size %#x",
			header.ImageDMAOffset,
			dataSize,
		)
	}
	if uint64(header.PaletteDMAOffset) >= dataSize {
		return fmt.Errorf(
			"palette DMA offset %#x is outside T32 data size %#x",
			header.PaletteDMAOffset,
			dataSize,
		)
	}
	return nil
}

func DecodeCSM1RGBA32(payload []byte) []RGBA {
	palette := make([]RGBA, 256)
	for logicalIndex := range palette {
		storedIndex := ClutPermutation(logicalIndex)
		position := storedIndex * 4
		palette[logicalIndex] = RGBA{
			R: payload[position+0],
			G: payload[position+1],
			B: payload[position+2],
			A: ExpandPS2Alpha(payload[position+3]),
		}
	}
	return palette
}
