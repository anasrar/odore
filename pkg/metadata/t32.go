package metadata

type PaletteT32 struct {
	Index uint64 `json:"index"`
	Path  string `json:"path"`
}

type TextureT32 struct {
	Index    uint64       `json:"index"`
	Palettes []PaletteT32 `json:"palettes"`
}

type MetadataT32 struct {
	Textures []TextureT32 `json:"textures"`
}
