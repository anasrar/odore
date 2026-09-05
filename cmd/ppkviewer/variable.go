package main

import rl "github.com/gen2brain/raylib-go/raylib"

var Version = "dev"

var (
	width      float32 = 600
	height     float32 = 500
	background         = [3]float32{0.071, 0.071, 0.071}
)

var (
	Input string
)

var camera = rl.NewCamera3D(
	rl.NewVector3(0, 2.8, 2.8),
	rl.NewVector3(0, 1.2, 0),
	rl.NewVector3(0, 1, 0),
	45,
	rl.CameraPerspective,
)

var (
	textures = []*Texture{}
)
