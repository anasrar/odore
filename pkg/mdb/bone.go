package mdb

type Bone struct {
	X        float32 `json:"x"`
	Y        float32 `json:"y"`
	Z        float32 `json:"z"`
	Parent   int16   `json:"parent"`
	Unknown0 int16   `json:"unknown0"`
}

type BoneContainer struct {
	Total   uint16 `json:"bone_total" skip:""`
	Entries []Bone `json:"entries" length:"Total"`
}

type BoneNode struct {
	Bone     Bone        `json:"bone"`
	Parent   *BoneNode   `json:"-"`
	Children []*BoneNode `json:"children"`
}
