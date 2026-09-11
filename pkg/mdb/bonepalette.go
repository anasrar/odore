package mdb

type BonePaletteOffset struct {
	Offset uint32 `json:"offset"`
	Total  uint32 `json:"bone_palette_offset_total"`
}

type BonePaletteHeader struct {
	Total   uint32              `json:"bone_palette_header_total"`
	Entries []BonePaletteOffset `json:"entries" length:"Total"`
}

type BonePalette struct {
	Total   uint32  `json:"bone_palette_total" skip:""`
	Entries []uint8 `json:"entries" length:"Total"`
}
