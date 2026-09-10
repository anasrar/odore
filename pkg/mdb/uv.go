package mdb

type UV struct {
	U float32 `json:"u"`
	V float32 `json:"v"`
	Q float32 `json:"q"`
	W float32 `json:"w"`
}

type UVContainer struct {
	Total   uint16 `json:"uv_total" skip:""`
	Entries []UV   `json:"entries" length:"Total"`
}
