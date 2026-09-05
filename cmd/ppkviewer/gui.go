package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/anasrar/odore/pkg/ppk"
	rlig "github.com/anasrar/odore/pkg/raylib_imgui"
	"github.com/anasrar/odore/pkg/t32"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func cleanUpTextures() {
	for _, entry := range textures {
		rl.UnloadTexture(entry.Texture)
		entry.Ref.Destroy()
	}
}

func drop(input string) error {
	file, err := os.Open(input)
	if err != nil {
		return err
	}
	defer file.Close()

	ppkContainer := ppk.New()
	if err := ppk.FromStream(ppkContainer, file); err != nil {
		return err
	}

	tmpTextures := []*Texture{}

	for t32Index, entry := range ppkContainer.T32Container.Entries {
		t32Container := t32.New()
		if err := t32.FromStreamWithOffset(t32Container, file, entry.Offset); err != nil {
			return err
		}

		for textureIndex, texture := range t32Container.Textures {
			for paletteIndex := range texture.Palettes {
				decoded, err := t32Container.DecodeTexture(textureIndex, paletteIndex)
				if err != nil {
					return err
				}

				filename := fmt.Sprintf("index_%03d_texture_%03d_palette_%03d.png", t32Index, textureIndex, paletteIndex)

				img := rl.NewImageFromImage(decoded.Image())
				defer rl.UnloadImage(img)
				tex := rl.LoadTextureFromImage(img)

				tmpTextures = append(tmpTextures, TextureNew(filename, tex))
			}
		}
	}

	cleanUpTextures()
	textures = tmpTextures

	return nil
}

func gui(input string) error {
	rl.InitWindow(int32(width), int32(height), "ppk Viewer")
	defer rl.CloseWindow()
	rl.SetTargetFPS(30)

	rlig.Load()
	defer rlig.Unload()

	defer cleanUpTextures()

	if input != "" {
		Input = input
		if err := drop(input); err != nil {
			log.Print(err)
		}
	}

	for !rl.WindowShouldClose() {
		rlig.Update()

		if rl.IsWindowResized() {
			width = float32(rl.GetScreenWidth())
			height = float32(rl.GetScreenHeight())
		}

		if rl.IsFileDropped() {
			filePath := rl.LoadDroppedFiles()[0]
			defer rl.UnloadDroppedFiles()

			Input = filePath
			if err := drop(filePath); err != nil {
				log.Println(err)
			}
		}

		if rl.IsKeyDown(rl.KeyW) {
			rl.CameraMoveForward(&camera, 1*rl.GetFrameTime(), 0)
		}
		if rl.IsKeyDown(rl.KeyS) {
			rl.CameraMoveForward(&camera, -1*rl.GetFrameTime(), 0)
		}

		if rl.IsKeyDown(rl.KeyA) {
			rl.CameraMoveRight(&camera, -1*rl.GetFrameTime(), 0)
		}
		if rl.IsKeyDown(rl.KeyD) {
			rl.CameraMoveRight(&camera, 1*rl.GetFrameTime(), 0)
		}

		if rl.IsKeyDown(rl.KeyQ) {
			rl.CameraMoveUp(&camera, -1*rl.GetFrameTime())
		}
		if rl.IsKeyDown(rl.KeyE) {
			rl.CameraMoveUp(&camera, 1*rl.GetFrameTime())
		}

		if rl.IsKeyDown(rl.KeyLeft) {
			rl.CameraYaw(&camera, 1*rl.GetFrameTime(), 0)
		}
		if rl.IsKeyDown(rl.KeyRight) {
			rl.CameraYaw(&camera, -1*rl.GetFrameTime(), 0)
		}

		if rl.IsKeyDown(rl.KeyUp) {
			rl.CameraPitch(&camera, 0.5*rl.GetFrameTime(), 0, 0, 0)
		}
		if rl.IsKeyDown(rl.KeyDown) {
			rl.CameraPitch(&camera, -0.5*rl.GetFrameTime(), 0, 0, 0)
		}

		imgui.NewFrame()

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, 12), imgui.CondAlways, imgui.NewVec2(1, 0))
		imgui.BeginV("View", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.ColorEdit3V("Background", &(background), imgui.ColorEditFlagsNoInputs)
		if imgui.Button("Reset View") {
			camera.Position = rl.NewVector3(0, 2.8, 2.8)
			camera.Target = rl.NewVector3(0, 1.2, 0)
		}
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, 104), imgui.CondAlways, imgui.NewVec2(1, 0))
		imgui.SetNextWindowSizeConstraints(imgui.NewVec2(108, 108), imgui.NewVec2(108, 216))
		imgui.BeginV("T32", nil, imgui.WindowFlagsNoMove)
		for _, tex := range textures {
			imgui.Image(tex.Ref, imgui.NewVec2(72, 72))
			if imgui.BeginItemTooltip() {
				imgui.Text(tex.Name)
				imgui.Image(tex.Ref, imgui.NewVec2(float32(tex.Texture.Width), float32(tex.Texture.Height)))
				imgui.EndTooltip()
			}
		}
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(12, height-12), imgui.CondAlways, imgui.NewVec2(0, 1))
		imgui.BeginV("Credit", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.Text(fmt.Sprintf("https://github.com/anasrar/odore@%s", Version))
		imgui.End()

		rl.BeginDrawing()
		rl.ClearBackground(
			rl.NewColor(
				uint8(background[0]*0xFF),
				uint8(background[1]*0xFF),
				uint8(background[2]*0xFF),
				0xFF,
			),
		)

		rl.BeginDrawing()
		rl.ClearBackground(
			rl.NewColor(
				uint8(background[0]*0xFF),
				uint8(background[1]*0xFF),
				uint8(background[2]*0xFF),
				0xFF,
			),
		)

		rl.BeginMode3D(camera)

		rl.DrawGrid(4, 0.5)

		rl.EndMode3D()

		rlig.Render()
		rl.EndDrawing()
	}

	return nil
}
