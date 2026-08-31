package t32

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/anasrar/odore/pkg/utils"
)

type sourceIdentity struct {
	raw          []byte
	header       Header
	textures     []Texture
	fingerprints [][][32]byte
}

type decodedProvenance struct {
	source       *sourceIdentity
	textureIndex int
	paletteIndex int
	fingerprint  [32]byte
}

func (c *Container) hydrateDecodedTextures() error {
	for textureIndex := range c.Textures {
		texture := &c.Textures[textureIndex]
		if !texture.IsIndexed() {
			continue
		}
		texture.Palettes = make([]DecodedTexture, int(texture.PaletteTotal))
		for paletteIndex := range texture.Palettes {
			decoded, err := c.decodeTextureFromDMA(textureIndex, paletteIndex)
			if err != nil {
				return fmt.Errorf(
					"decode texture %d palette %d: %w",
					textureIndex,
					paletteIndex,
					err,
				)
			}
			texture.Palettes[paletteIndex] = *decoded
		}
	}
	return nil
}

func (c *Container) captureSource(stream io.ReadSeeker, fileSize, baseOffset uint64) error {
	end := c.serializedDataEnd()
	if end > fileSize-baseOffset {
		return fmt.Errorf(
			"T32 data ends at %#x, stream only has %#x bytes from its base",
			end,
			fileSize-baseOffset,
		)
	}

	raw := make([]byte, end)
	if err := utils.SeekAbsolute(stream, baseOffset); err != nil {
		return err
	}
	if _, err := io.ReadFull(stream, raw); err != nil {
		return fmt.Errorf("read original T32 data: %w", err)
	}

	source := &sourceIdentity{
		raw:          raw,
		header:       c.Header,
		textures:     make([]Texture, len(c.Textures)),
		fingerprints: make([][][32]byte, len(c.Textures)),
	}
	for textureIndex := range c.Textures {
		source.textures[textureIndex] = c.Textures[textureIndex]
		source.textures[textureIndex].Palettes = nil
		source.fingerprints[textureIndex] = make(
			[][32]byte,
			len(c.Textures[textureIndex].Palettes),
		)
		for paletteIndex := range c.Textures[textureIndex].Palettes {
			decoded := &c.Textures[textureIndex].Palettes[paletteIndex]
			fingerprint := fingerprintDecoded(decoded)
			source.fingerprints[textureIndex][paletteIndex] = fingerprint
			decoded.provenance = &decodedProvenance{
				source:       source,
				textureIndex: textureIndex,
				paletteIndex: paletteIndex,
				fingerprint:  fingerprint,
			}
		}
	}

	c.source = source
	c.sourceExact = true
	return nil
}

func (c *Container) serializedDataEnd() uint64 {
	end := uint64(HeaderSize) + uint64(len(c.Textures))*TextureEntrySize
	for _, chain := range []DMAChain{c.ImageDMA, c.PaletteDMA} {
		chainEnd := uint64(chain.Offset) + uint64(len(chain.Transfers)+1)*DMATagSize
		if chainEnd > end {
			end = chainEnd
		}
		for i := range chain.Transfers {
			transferEnd := chain.Transfers[i].PacketOffset +
				UploadHeaderSize + uint64(len(chain.Transfers[i].Data))
			if transferEnd > end {
				end = transferEnd
			}
		}
	}
	return end
}

func fingerprintDecoded(decoded *DecodedTexture) [32]byte {
	hash := sha256.New()
	var scalar [8]byte
	binary.LittleEndian.PutUint64(scalar[:], uint64(decoded.Width))
	hash.Write(scalar[:])
	binary.LittleEndian.PutUint64(scalar[:], uint64(decoded.Height))
	hash.Write(scalar[:])
	hash.Write([]byte{byte(decoded.Format)})
	hash.Write(decoded.Indices)
	for _, entry := range decoded.Palette {
		hash.Write([]byte{entry.R, entry.G, entry.B, entry.A})
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func (c *Container) matchesSource() bool {
	if c == nil || !c.sourceExact || c.source == nil {
		return false
	}
	if c.Header != c.source.header || len(c.Textures) != len(c.source.textures) {
		return false
	}
	for textureIndex := range c.Textures {
		texture := &c.Textures[textureIndex]
		if !sameTextureDescriptor(texture, &c.source.textures[textureIndex]) {
			return false
		}
		if len(texture.Palettes) != len(c.source.fingerprints[textureIndex]) {
			return false
		}
		for paletteIndex := range texture.Palettes {
			if fingerprintDecoded(&texture.Palettes[paletteIndex]) !=
				c.source.fingerprints[textureIndex][paletteIndex] {
				return false
			}
		}
	}
	return true
}

func sameTextureDescriptor(left, right *Texture) bool {
	return left.PaletteTotal == right.PaletteTotal &&
		left.Reserved0 == right.Reserved0 &&
		left.GsTex0Raw == right.GsTex0Raw &&
		left.GsTex1Raw == right.GsTex1Raw &&
		left.Reserved1 == right.Reserved1 &&
		left.NextTextureIndex == right.NextTextureIndex &&
		left.NextTextureSize == right.NextTextureSize &&
		left.Index == right.Index &&
		left.DataSize == right.DataSize &&
		left.GsTex0 == right.GsTex0 &&
		left.GsTex1 == right.GsTex1
}
