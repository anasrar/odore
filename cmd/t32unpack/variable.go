package main

import rl "github.com/gen2brain/raylib-go/raylib"

var Version = "dev"

var (
	width        float32 = 600
	height       float32 = 500
	background           = [3]float32{0.071, 0.071, 0.071}
	zoomDeadZone         = rl.NewRectangle(width-74, 58, 64, height-108)
)

var (
	Input  string
	Offset uint32
	IsGUI  bool
)

var (
	camera       = rl.NewCamera2D(rl.Vector2Zero(), rl.Vector2Zero(), 0, 1)
	matrix       = rl.MatrixIdentity()
	entries      = []*Entry{}
	currentEntry = -1
	canConvert   = false
)
