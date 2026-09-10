package mdb

type Normal struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
	W float32 `json:"w"`
}

type NormalContainer struct {
	Total   uint16   `json:"normal_total" skip:""`
	Entries []Normal `json:"entries" length:"Total"`
}
