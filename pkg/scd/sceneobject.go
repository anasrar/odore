package scd

import "github.com/anasrar/odore/pkg/utils"

type SceneObject struct {
	Position  utils.Vector3 `json:"position"`
	Rotation  utils.Vector3 `json:"rotation"`
	Scale     utils.Vector3 `json:"scale"`
	Index     uint32        `json:"index "`
	Unknown0  uint32        `json:"unknown0"`
	Unknown1  uint32        `json:"unknown1"`
	Unknown2  uint32        `json:"unknown2"`
	Unknown3  uint32        `json:"unknown3"`
	Unknown4  uint32        `json:"unknown4"`
	Unknown5  uint32        `json:"unknown5"`
	Unknown6  uint32        `json:"unknown6"`
	Unknown7  uint32        `json:"unknown7"`
	MDBOffset uint32        `json:"mdb_offset"`
	Unknown8  uint32        `json:"unknown8"`
}
