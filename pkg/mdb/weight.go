package mdb

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Weight struct {
	Joint0  uint16  `json:"joint0"`
	Joint1  uint16  `json:"joint1"`
	Joint2  uint16  `json:"joint2"`
	Joint3  uint16  `json:"joint3"`
	Weight0 float32 `json:"weight0"`
	Weight1 float32 `json:"weight1"`
	Weight2 float32 `json:"weight2"`
	Weight3 float32 `json:"weight3"`
}

type WeightContainer struct {
	Total   uint16   `json:"weight_total" skip:""`
	Entries []Weight `json:"entries"`
}

type weight3Joint struct {
	Unknown0 uint32
	Joint0   uint32
	Joint1   uint32
	Joint2   uint32
	Weight0  float32
	Weight1  float32
	Weight2  float32
	Unknown1 float32
}

type weight2Joint struct {
	Joint0  uint32
	Joint1  uint32
	Weight0 float32
	Weight1 float32
}

func resolveJoint(joint uint32, palette []uint8) (uint16, error) {
	if joint%4 != 0 {
		return 0, fmt.Errorf("joint value %#x is not aligned to 4 bytes", joint)
	}

	index := joint / 4
	if index >= uint32(len(palette)) {
		return 0, fmt.Errorf("joint index %d is outside bone palette length %d", index, len(palette))
	}

	return uint16(palette[index]), nil
}

func (c *WeightContainer) unmarshal(stream io.Reader, skinningFormat uint32, palette []uint8) error {
	c.Entries = make([]Weight, c.Total)

	for i := range c.Entries {
		weight := &c.Entries[i]

		switch skinningFormat {
		case SkinningFormat0401:
			var entry weight3Joint
			if err := binary.Read(stream, binary.LittleEndian, &entry); err != nil {
				return err
			}

			joints := [...]uint32{entry.Joint0, entry.Joint1, entry.Joint2}
			resolved := [3]uint16{}
			for j, joint := range joints {
				index, err := resolveJoint(joint, palette)
				if err != nil {
					return err
				}
				resolved[j] = index
			}

			weight.Joint0 = resolved[0]
			weight.Joint1 = resolved[1]
			weight.Joint2 = resolved[2]
			weight.Weight0 = entry.Weight0
			weight.Weight1 = entry.Weight1
			weight.Weight2 = entry.Weight2
		case SkinningFormat0500, SkinningFormat0501:
			var entry weight2Joint
			if err := binary.Read(stream, binary.LittleEndian, &entry); err != nil {
				return err
			}

			joint0, err := resolveJoint(entry.Joint0, palette)
			if err != nil {
				return err
			}
			joint1, err := resolveJoint(entry.Joint1, palette)
			if err != nil {
				return err
			}

			weight.Joint0 = joint0
			weight.Joint1 = joint1
			weight.Weight0 = entry.Weight0
			weight.Weight1 = entry.Weight1
		default:
			return fmt.Errorf("unsupported skinning format %#x", skinningFormat)
		}
	}

	return nil
}
