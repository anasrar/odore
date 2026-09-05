package main

import (
	"github.com/AllenDang/cimgui-go/imgui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Texture struct {
	Name    string
	Texture rl.Texture2D
	Ref     imgui.TextureRef
}

func TextureNew(name string, texture rl.Texture2D) *Texture {
	return &Texture{
		Name:    name,
		Texture: texture,
		Ref:     *imgui.NewTextureRefTextureID(imgui.TextureID(texture.ID)),
	}
}
