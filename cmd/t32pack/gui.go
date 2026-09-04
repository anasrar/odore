package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/anasrar/odore/pkg/metadata"
	rlig "github.com/anasrar/odore/pkg/raylib_imgui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func drop(input string) error {
	metadataBuf, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	m := metadata.MetadataT32{}
	if err := json.Unmarshal(metadataBuf, &m); err != nil {
		return err
	}

	for _, entry := range entries {
		rl.UnloadTexture(entry.Texture)
		entry.TextureRef.Destroy()
	}

	entries = []*Entry{}

	for textureIndex, texture := range m.Textures {
		for paletteIndex, palette := range texture.Palettes {
			filename := fmt.Sprintf("texture_%03d_palette_%03d.png", textureIndex, paletteIndex)
			pngFile, err := os.Open(palette.Path)
			if err != nil {
				return err
			}
			defer pngFile.Close()

			img, err := png.Decode(pngFile)
			if err != nil {
				return err
			}

			if _, ok := img.(*image.Paletted); !ok {
				return fmt.Errorf("not indexed PNG")
			}

			rlImg := rl.NewImageFromImage(img)
			texture := rl.LoadTextureFromImage(rlImg)
			rl.UnloadImage(rlImg)

			entries = append(
				entries,
				&Entry{
					Name:       filename,
					Texture:    texture,
					TextureRef: *imgui.NewTextureRefTextureID(imgui.TextureID(texture.ID)),
				},
			)

		}
	}

	entry := entries[0]
	currentEntry = 0

	matrix = rl.MatrixTranslate(
		(width/2)-(float32(entry.Texture.Width)/2),
		(height/2)-(float32(entry.Texture.Height)/2),
		0,
	)

	canConvert = true

	return nil
}

func zoom(wheel float32) {
	isInsidePreview := rl.CheckCollisionPointRec(rl.GetMousePosition(), zoomDeadZone)

	if isInsidePreview {
		return
	}

	scale := float32(0)
	switch wheel {
	case 1:
		scale = 6.0 / 5.0
	case -1:
		scale = 5.0 / 6.0
	}
	positionX := rl.GetMousePosition().X
	positionY := rl.GetMousePosition().Y
	matrix = rl.MatrixMultiply(
		matrix,
		rl.MatrixTranslate(-positionX, -positionY, 0),
	)
	matrix = rl.MatrixMultiply(
		matrix,
		rl.MatrixScale(scale, scale, 1),
	)
	matrix = rl.MatrixMultiply(
		matrix,
		rl.MatrixTranslate(positionX, positionY, 0),
	)
}

func gui(input string) error {
	rl.InitWindow(int32(width), int32(height), "T32 Unpack")
	defer rl.CloseWindow()
	rl.SetTargetFPS(30)

	rlig.Load()
	defer rlig.Unload()

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

		if rl.IsMouseButtonDown(1) {
			positionX := rl.GetMouseDelta().X
			positionY := rl.GetMouseDelta().Y
			matrix = rl.MatrixMultiply(
				matrix,
				rl.MatrixTranslate(positionX, positionY, 0),
			)
		}

		wheel := rl.GetMouseWheelMoveV().Y
		if wheel != 0 {
			zoom(wheel)
		}

		imgui.NewFrame()

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, 12), imgui.CondAlways, imgui.NewVec2(1, 0))
		imgui.BeginV("View", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.ColorEdit3V("Background", &(background), imgui.ColorEditFlagsNoInputs)
		if imgui.Button("Reset View") {
			if currentEntry != -1 {
				entry := entries[currentEntry]
				matrix = rl.MatrixTranslate(
					(width/2)-(float32(entry.Texture.Width)/2),
					(height/2)-(float32(entry.Texture.Height)/2),
					0,
				)
			}
		}
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, 104), imgui.CondAlways, imgui.NewVec2(1, 0))
		imgui.SetNextWindowSizeConstraints(imgui.NewVec2(84, 108), imgui.NewVec2(84, 216))
		imgui.BeginV("Entries", nil, imgui.WindowFlagsNoMove)
		for i, entry := range entries {
			if imgui.ImageButton(entry.Name, entry.TextureRef, imgui.NewVec2(42, 42)) {
				currentEntry = i
				matrix = rl.MatrixTranslate(
					(width/2)-(float32(entry.Texture.Width)/2),
					(height/2)-(float32(entry.Texture.Height)/2),
					0,
				)
			}
		}

		{
			windowEntriesRect := imgui.InternalCurrentWindow().Size()
			zoomDeadZoneMin := imgui.InternalCurrentWindow().Pos()
			zoomDeadZone.X = zoomDeadZoneMin.X
			zoomDeadZone.Y = zoomDeadZoneMin.Y
			zoomDeadZone.Width = windowEntriesRect.X
			zoomDeadZone.Height = windowEntriesRect.Y
		}

		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, height-12), imgui.CondAlways, imgui.NewVec2(1, 1))
		imgui.BeginV("Actions", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
		imgui.BeginDisabledV(!canConvert)
		if imgui.Button("Pack") {
			go func() {
				if err := pack(Input); err != nil {
					log.Print(err)
				}
			}()
		}
		imgui.EndDisabled()
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

		if currentEntry != -1 {
			entry := entries[currentEntry]

			rl.BeginMode2D(camera)

			translate := rl.NewVector3(0, 0, 0)
			rotation := rl.NewQuaternion(0, 0, 0, 1)
			scale := rl.NewVector3(1, 1, 1)

			rl.MatrixDecompose(matrix, &translate, &rotation, &scale)

			rl.DrawRectangleLinesEx(rl.NewRectangle(translate.X, translate.Y, float32(entry.Texture.Width)*scale.X, float32(entry.Texture.Height)*scale.Y), 1, rl.Gray)

			rl.PushMatrix()
			rl.Translatef(translate.X, translate.Y, 0)
			rl.Scalef(scale.X, scale.Y, 1)

			rl.DrawTexture(entry.Texture, 0, 0, rl.White)

			rl.PopMatrix()

			rl.EndMode2D()
		}

		rlig.Render()
		rl.EndDrawing()
	}

	return nil
}
