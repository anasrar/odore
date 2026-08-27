package manifest

type Palette struct {
	Index uint64 `json:"index"`
	Path  string `json:"path"`
}

type Texture struct {
	Index    uint64    `json:"index"`
	Palettes []Palette `json:"palettes"`
}

type Metadata struct {
	Textures []Texture `json:"textures"`
}
