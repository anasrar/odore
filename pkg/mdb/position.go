package mdb

type Position struct {
	X    float32 `json:"x"`
	Y    float32 `json:"y"`
	Z    float32 `json:"z"`
	Flag uint32  `json:"flag"`
}

type PositionContainer struct {
	Total   uint16     `json:"vertex_total" skip:""`
	Entries []Position `json:"entries" length:"Total"`
}
