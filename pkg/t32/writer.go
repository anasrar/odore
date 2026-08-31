package t32

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"

	"github.com/anasrar/graphicsynthesizer"
)

const (
	packetAlignment      = 0x80
	maximumImageDataSize = 0x7e000
	gsPageSize           = 0x2000
	gsBlockSize          = 0x100
)

type textureDescriptor struct {
	PaletteTotal     uint32
	Reserved0        uint32
	GsTex0Raw        uint64
	GsTex1Raw        uint32
	Reserved1        uint32
	NextTextureIndex uint32
	NextTextureSize  uint32
}

type textureWritePlan struct {
	texture      *Texture
	format       PixelStorageMode
	width        int
	height       int
	tbw          int
	tw           uint8
	th           uint8
	base         int
	dataSize     int
	paletteStart int
	descriptor   textureDescriptor
}

type uploadWritePlan struct {
	packetOffset int
	base         int
	width        int
	height       int
	data         []byte
}

func (c *Container) Write(writer io.Writer) error {
	if c == nil {
		return ErrContainerIsNil
	}
	if writer == nil {
		return errors.New("writer is nil")
	}
	if c.matchesSource() {
		return writeFull(writer, c.source.raw)
	}

	data, err := c.marshalCanonical()
	if err != nil {
		return err
	}
	return writeFull(writer, data)
}

func (c *Container) marshalCanonical() ([]byte, error) {
	texturePlans, imageData, paletteTotal, err := c.buildTextureWritePlans()
	if err != nil {
		return nil, err
	}
	imageUploads := splitImageUploads(imageData)
	if len(imageUploads) > 0xffff || paletteTotal > 0x4000 {
		return nil, errors.New("T32 has too many DMA transfers")
	}

	imageChainOffset := HeaderSize + len(texturePlans)*TextureEntrySize
	packetCursor := align(
		imageChainOffset+(len(imageUploads)+1)*DMATagSize,
		packetAlignment,
	)
	for index := range imageUploads {
		imageUploads[index].packetOffset = packetCursor
		packetCursor = align(
			packetCursor+UploadHeaderSize+len(imageUploads[index].data),
			packetAlignment,
		)
	}

	paletteChainOffset := packetCursor
	paletteUploads := make([]uploadWritePlan, 0, paletteTotal)
	packetCursor = align(
		paletteChainOffset+(paletteTotal+1)*DMATagSize,
		packetAlignment,
	)
	for _, plan := range texturePlans {
		for paletteIndex := range plan.texture.Palettes {
			paletteUploads = append(paletteUploads, uploadWritePlan{
				packetOffset: packetCursor,
				base:         len(paletteUploads),
				width:        16,
				height:       16,
				data: encodePalette(
					plan.texture.Palettes[paletteIndex].Palette,
				),
			})
			packetCursor = align(
				packetCursor+UploadHeaderSize+0x400,
				packetAlignment,
			)
		}
	}
	if len(paletteUploads) == 0 {
		return nil, errors.New("T32 must contain at least one palette")
	}
	fileSize := paletteUploads[len(paletteUploads)-1].packetOffset +
		UploadHeaderSize + len(paletteUploads[len(paletteUploads)-1].data)
	data := make([]byte, fileSize)

	header := Header{
		TextureTotal:       uint32(len(texturePlans)),
		TransferWidth:      64,
		TransferHeight:     uint16(len(imageData) / (64 * 4)),
		ImageDMAOffset:     uint32(imageChainOffset),
		PaletteDMAOffset:   uint32(paletteChainOffset),
		ImagePacketTotal:   uint16(len(imageUploads)),
		PalettePacketTotal: uint16(len(paletteUploads)),
		Reserved0:          c.Reserved0,
		Reserved1:          c.Reserved1,
		FirstTextureSize:   uint32(texturePlans[0].dataSize),
	}
	if err := writeBinaryAt(data, 0, header); err != nil {
		return nil, err
	}
	for index := range texturePlans {
		if err := writeBinaryAt(
			data,
			HeaderSize+index*TextureEntrySize,
			texturePlans[index].descriptor,
		); err != nil {
			return nil, err
		}
	}
	if err := writeUploadChain(data, imageChainOffset, imageUploads); err != nil {
		return nil, fmt.Errorf("write image DMA: %w", err)
	}
	if err := writeUploadChain(data, paletteChainOffset, paletteUploads); err != nil {
		return nil, fmt.Errorf("write palette DMA: %w", err)
	}
	return data, nil
}

func (c *Container) buildTextureWritePlans() ([]textureWritePlan, []byte, int, error) {
	if len(c.Textures) == 0 {
		return nil, nil, 0, errors.New("T32 must contain at least one texture")
	}

	plans := make([]textureWritePlan, len(c.Textures))
	dataSize := 0
	paletteTotal := 0
	for textureIndex := range c.Textures {
		texture := &c.Textures[textureIndex]
		if len(texture.Palettes) == 0 {
			return nil, nil, 0, fmt.Errorf("texture %d has no palettes", textureIndex)
		}
		base := &texture.Palettes[0]
		if err := validateDecodedTexture(base); err != nil {
			return nil, nil, 0, fmt.Errorf("texture %d: %w", textureIndex, err)
		}
		format := base.Format
		if format != PSMT4 && format != PSMT8 {
			format = inferDecodedPixelStorageMode(base)
		}
		paletteLimit := 256
		pageHeight := 64
		if format == PSMT4 {
			paletteLimit = 16
			pageHeight = 128
		}
		for paletteIndex := range texture.Palettes {
			palette := &texture.Palettes[paletteIndex]
			if palette.Width != base.Width || palette.Height != base.Height {
				return nil, nil, 0, fmt.Errorf(
					"texture %d palette %d dimensions are %dx%d, expected %dx%d",
					textureIndex,
					paletteIndex,
					palette.Width,
					palette.Height,
					base.Width,
					base.Height,
				)
			}
			if len(palette.Palette) == 0 || len(palette.Palette) > paletteLimit {
				return nil, nil, 0, fmt.Errorf(
					"texture %d palette %d has %d entries, %s allows 1..%d",
					textureIndex,
					paletteIndex,
					len(palette.Palette),
					format,
					paletteLimit,
				)
			}
			for pixel, value := range base.Indices {
				if int(value) >= len(palette.Palette) {
					return nil, nil, 0, fmt.Errorf(
						"texture %d palette %d: pixel %d uses index %d outside %d entries",
						textureIndex,
						paletteIndex,
						pixel,
						value,
						len(palette.Palette),
					)
				}
			}
		}

		pagesWide := (base.Width + 127) / 128
		pagesHigh := (base.Height + pageHeight - 1) / pageHeight
		textureDataSize := pagesWide * pagesHigh * gsPageSize
		if dataSize/gsBlockSize > 0x3fff {
			return nil, nil, 0, fmt.Errorf("texture %d exceeds GS memory", textureIndex)
		}
		tw := uint8(bits.TrailingZeros(uint(base.Width)))
		th := uint8(bits.TrailingZeros(uint(base.Height)))
		plan := textureWritePlan{
			texture:      texture,
			format:       format,
			width:        base.Width,
			height:       base.Height,
			tbw:          pagesWide * 2,
			tw:           tw,
			th:           th,
			base:         dataSize / gsBlockSize,
			dataSize:     textureDataSize,
			paletteStart: paletteTotal,
		}
		if plan.paletteStart > 0x3fff {
			return nil, nil, 0, fmt.Errorf(
				"texture %d palette base %#x exceeds TEX0 CBP",
				textureIndex,
				plan.paletteStart,
			)
		}
		plans[textureIndex] = plan
		dataSize += textureDataSize
		paletteTotal += len(texture.Palettes)
	}
	if dataSize > len(graphicsynthesizer.GsMem) || dataSize/gsBlockSize > 0x4000 {
		return nil, nil, 0, fmt.Errorf("texture data size %#x exceeds GS memory", dataSize)
	}
	if dataSize/(64*4) > 0xffff {
		return nil, nil, 0, fmt.Errorf("texture transfer height %#x exceeds uint16", dataSize/(64*4))
	}

	gsMutex.Lock()
	clear(graphicsynthesizer.GsMem)
	for index := range plans {
		plan := &plans[index]
		indices := plan.texture.Palettes[0].Indices
		switch plan.format {
		case PSMT8:
			graphicsynthesizer.WriteTexPSMT8(
				plan.base,
				plan.tbw,
				0,
				0,
				plan.width,
				plan.height,
				indices,
			)
		case PSMT4:
			packed := make([]byte, (len(indices)+1)/2)
			for pixel, value := range indices {
				if pixel&1 == 0 {
					packed[pixel/2] = value & 0x0f
				} else {
					packed[pixel/2] |= value << 4
				}
			}
			graphicsynthesizer.WriteTexPSMT4(
				plan.base,
				plan.tbw,
				0,
				0,
				plan.width,
				plan.height,
				packed,
			)
		}
	}
	imageData := make([]byte, dataSize)
	graphicsynthesizer.ReadTexPSMCT32(
		0,
		1,
		0,
		0,
		64,
		dataSize/(64*4),
		imageData,
	)
	gsMutex.Unlock()

	for index := range plans {
		plan := &plans[index]
		tex0Raw := uint64(plan.base&0x3fff) |
			uint64(plan.tbw&0x3f)<<14 |
			uint64(plan.format&0x3f)<<20 |
			uint64(plan.tw&0x0f)<<26 |
			uint64(plan.th&0x0f)<<30 |
			uint64(plan.paletteStart&0x3fff)<<37 |
			uint64(1)<<61
		descriptor := textureDescriptor{
			PaletteTotal: uint32(len(plan.texture.Palettes)),
			Reserved0:    plan.texture.Reserved0,
			GsTex0Raw:    tex0Raw,
			GsTex1Raw:    0x260,
			Reserved1:    plan.texture.Reserved1,
		}
		if index+1 < len(plans) {
			descriptor.NextTextureIndex = uint32(index + 1)
			descriptor.NextTextureSize = uint32(plans[index+1].dataSize)
		}
		plan.descriptor = descriptor
	}
	return plans, imageData, paletteTotal, nil
}

func validateDecodedTexture(decoded *DecodedTexture) error {
	if decoded.Width <= 0 || decoded.Height <= 0 ||
		decoded.Width&(decoded.Width-1) != 0 || decoded.Height&(decoded.Height-1) != 0 {
		return fmt.Errorf(
			"dimensions %dx%d must be positive powers of two",
			decoded.Width,
			decoded.Height,
		)
	}
	if decoded.Width > 1<<15 || decoded.Height > 1<<15 {
		return fmt.Errorf("dimensions %dx%d exceed TEX0", decoded.Width, decoded.Height)
	}
	if len(decoded.Indices) != decoded.Width*decoded.Height {
		return fmt.Errorf(
			"has %d indices, expected %d",
			len(decoded.Indices),
			decoded.Width*decoded.Height,
		)
	}
	return nil
}

func inferDecodedPixelStorageMode(decoded *DecodedTexture) PixelStorageMode {
	if len(decoded.Palette) <= 16 {
		for _, value := range decoded.Indices {
			if value > 0x0f {
				return PSMT8
			}
		}
		return PSMT4
	}
	return PSMT8
}

func splitImageUploads(imageData []byte) []uploadWritePlan {
	uploads := make([]uploadWritePlan, 0, (len(imageData)+maximumImageDataSize-1)/maximumImageDataSize)
	for offset := 0; offset < len(imageData); {
		size := min(maximumImageDataSize, len(imageData)-offset)
		uploads = append(uploads, uploadWritePlan{
			base:   offset / gsBlockSize,
			width:  64,
			height: size / (64 * 4),
			data:   imageData[offset : offset+size],
		})
		offset += size
	}
	return uploads
}

func writeUploadChain(data []byte, chainOffset int, uploads []uploadWritePlan) error {
	for index := range uploads {
		upload := &uploads[index]
		qwc := len(upload.data) / 16
		address := upload.packetOffset - chainOffset
		if qwc > 0x7fff || address < 0 || uint64(address) > 0x7fffffff {
			return fmt.Errorf("upload %d cannot be represented by a DMA tag", index)
		}
		tag := DMATag{
			Word0: uint64(qwc+6) |
				uint64(DMAIDReference)<<28 |
				uint64(address)<<32,
		}
		if err := writeBinaryAt(data, chainOffset+index*DMATagSize, tag); err != nil {
			return err
		}
		packet := newUploadPacket(upload.base, upload.width, upload.height, qwc)
		if err := writeBinaryAt(data, upload.packetOffset, packet); err != nil {
			return err
		}
		copy(data[upload.packetOffset+UploadHeaderSize:], upload.data)
	}
	return writeBinaryAt(
		data,
		chainOffset+len(uploads)*DMATagSize,
		DMATag{Word0: uint64(DMAIDEnd) << 28},
	)
}

func newUploadPacket(base, width, height, qwc int) GIFUploadPacket {
	return GIFUploadPacket{
		SetupTag: GIFTag{
			Low:  uint64(1)<<60 | 4,
			High: 0x0e,
		},
		BITBLTBUF: GIFAD{
			Value:    uint64(base&0x3fff)<<32 | uint64(1)<<48,
			Register: GIFRegisterBITBLTBUF,
		},
		TRXPOS: GIFAD{
			Register: GIFRegisterTRXPOS,
		},
		TRXREG: GIFAD{
			Value:    uint64(width&0xfff) | uint64(height&0xfff)<<32,
			Register: GIFRegisterTRXREG,
		},
		TRXDIR: GIFAD{
			Register: GIFRegisterTRXDIR,
		},
		ImageTag: GIFTag{
			Low: uint64(qwc&0x7fff) | uint64(1)<<15 | uint64(2)<<58,
		},
	}
}

func encodePalette(palette []RGBA) []byte {
	data := make([]byte, 256*4)
	for logicalIndex := 0; logicalIndex < len(palette) && logicalIndex < 256; logicalIndex++ {
		storedIndex := ClutPermutation(logicalIndex)
		offset := storedIndex * 4
		entry := palette[logicalIndex]
		data[offset+0] = entry.R
		data[offset+1] = entry.G
		data[offset+2] = entry.B
		data[offset+3] = collapsePS2Alpha(entry.A)
	}
	return data
}

func collapsePS2Alpha(alpha uint8) uint8 {
	if alpha == 0xff {
		return 0x80
	}
	return (alpha + 1) / 2
}

func writeBinaryAt(destination []byte, offset int, value any) error {
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.LittleEndian, value); err != nil {
		return err
	}
	if offset < 0 || offset+buffer.Len() > len(destination) {
		return io.ErrShortBuffer
	}
	copy(destination[offset:], buffer.Bytes())
	return nil
}

func writeFull(writer io.Writer, data []byte) error {
	for len(data) != 0 {
		written, err := writer.Write(data)
		if err != nil {
			return err
		}
		if written <= 0 || written > len(data) {
			return io.ErrShortWrite
		}
		data = data[written:]
	}
	return nil
}

func align(value, alignment int) int {
	return (value + alignment - 1) &^ (alignment - 1)
}
