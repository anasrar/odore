package main

import (
	"github.com/anasrar/odore/pkg/mdb"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawBone(bone *mdb.BoneNode, scale float32) {
	rl.PushMatrix()
	rl.Translatef(bone.Bone.X*scale, bone.Bone.Y*scale, bone.Bone.Z*scale)

	rl.DrawSphere(
		rl.Vector3Zero(),
		0.02,
		rl.Green,
	)

	for _, child := range bone.Children {
		drawBone(child, scale)
	}

	rl.PopMatrix()
}
