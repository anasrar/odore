package mdb

type VertexBufferPosition struct {
	X    float32 `json:"x"`
	Y    float32 `json:"y"`
	Z    float32 `json:"z"`
	Flag uint32  `json:"flag"`
}

type VertexBufferContainerPosition struct {
	Total   uint16                 `json:"vertex_total" skip:""`
	Entries []VertexBufferPosition `json:"entries" length:"Total"`
}

type VertexBufferHeader struct {
	PositionOffset uint32 `json:"position_offset"`
	NormalOffset   uint32 `json:"normal_offset"`
	UVOffset       uint32 `json:"uv_offset"`
	ColorOffset    uint32 `json:"color_offset"`
	WeightOffset   uint32 `json:"weight_offset"`
	VertexTotal    uint16 `json:"vertex_total"`
	Unknown0       uint8  `json:"unknown0"`
	Material       uint8  `json:"material"`
}

type VertexBufferContainer struct {
	Header             VertexBufferHeader            `json:"header"`
	ContainerPositions VertexBufferContainerPosition `json:"positions"`
}

func NewVertexBufferContainer() *VertexBufferContainer {
	return &VertexBufferContainer{}
}
