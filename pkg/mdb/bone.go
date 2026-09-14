package mdb

import (
	"fmt"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

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
	Index    int         `json:"index"`
	Bone     Bone        `json:"bone"`
	Parent   *BoneNode   `json:"-"`
	Children []*BoneNode `json:"children"`
}

func (b *BoneNode) ConvertToGLTFSkin(doc *gltf.Document, mdbIndex int, boneTotal uint16, scale float32) (int, error) {
	if boneTotal == 0 {
		return 0, fmt.Errorf("MDB %d has no bones", mdbIndex)
	}

	joints := make([]int, int(boneTotal)+1)
	inverseBindMatrices := make([][4][4]float32, len(joints))
	for i := range joints {
		joints[i] = -1
	}

	var convert func(*BoneNode, [3]float64) (int, error)
	convert = func(bone *BoneNode, position [3]float64) (int, error) {
		if bone == nil {
			return 0, fmt.Errorf("MDB %d bone tree contains a nil bone", mdbIndex)
		}

		boneIndex := bone.Index
		name := fmt.Sprintf("mdb_%03d_bone_%03d", mdbIndex, boneIndex)
		if bone == b {
			boneIndex = int(boneTotal)
			name = fmt.Sprintf("mdb_%03d_bone_root", mdbIndex)
		} else if boneIndex < 0 || boneIndex >= int(boneTotal) {
			return 0, fmt.Errorf("MDB %d bone index %d is outside bone total %d", mdbIndex, boneIndex, boneTotal)
		}
		if joints[boneIndex] != -1 {
			return 0, fmt.Errorf("MDB %d bone tree repeats bone index %d", mdbIndex, boneIndex)
		}

		node := &gltf.Node{
			Name: name,
			Translation: [3]float64{
				float64(bone.Bone.X * scale),
				float64(bone.Bone.Y * scale),
				float64(bone.Bone.Z * scale),
			},
		}
		nodeIndex := len(doc.Nodes)
		doc.Nodes = append(doc.Nodes, node)
		joints[boneIndex] = nodeIndex

		position[0] += node.Translation[0]
		position[1] += node.Translation[1]
		position[2] += node.Translation[2]

		inverseBindMatrices[boneIndex] = [4][4]float32{
			{1, 0, 0, -float32(position[0])},
			{0, 1, 0, -float32(position[1])},
			{0, 0, 1, -float32(position[2])},
			{0, 0, 0, 1},
		}

		for _, child := range bone.Children {
			childIndex, err := convert(child, position)
			if err != nil {
				return 0, err
			}
			node.Children = append(node.Children, childIndex)
		}

		return nodeIndex, nil
	}

	rootIndex, err := convert(b, [3]float64{})
	if err != nil {
		return 0, err
	}
	for boneIndex, joint := range joints {
		if joint == -1 {
			return 0, fmt.Errorf("MDB %d bone tree is missing bone index %d", mdbIndex, boneIndex)
		}
	}

	inverseBindMatricesIndex := modeler.WriteAccessor(doc, gltf.TargetNone, inverseBindMatrices)
	doc.Skins = append(doc.Skins, &gltf.Skin{
		Name:                fmt.Sprintf("mdb_%03d_skeleton", mdbIndex),
		Skeleton:            gltf.Index(rootIndex),
		Joints:              joints,
		InverseBindMatrices: gltf.Index(inverseBindMatricesIndex),
	})

	return len(doc.Skins) - 1, nil
}
