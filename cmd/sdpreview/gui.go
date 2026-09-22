package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AllenDang/cimgui-go/imgui"
	rlig "github.com/anasrar/odore/pkg/raylib_imgui"
	"github.com/anasrar/odore/pkg/sd"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func drop(input string) error {
	file, err := os.Open(input)
	if err != nil {
		return err
	}
	defer file.Close()

	sdContainer := sd.New()
	if err := sd.FromStream(sdContainer, file); err != nil {
		return err
	}

	container = sdContainer

	return nil
}

func gui(input string) error {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(int32(width), int32(height), "SD Viewer")
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

		if rl.IsKeyDown(rl.KeyW) {
			rl.CameraMoveForward(&camera, cameraMoveSpeed*rl.GetFrameTime(), 0)
		}
		if rl.IsKeyDown(rl.KeyS) {
			rl.CameraMoveForward(&camera, -cameraMoveSpeed*rl.GetFrameTime(), 0)
		}

		if rl.IsKeyDown(rl.KeyA) {
			rl.CameraMoveRight(&camera, -cameraMoveSpeed*rl.GetFrameTime(), 0)
		}
		if rl.IsKeyDown(rl.KeyD) {
			rl.CameraMoveRight(&camera, cameraMoveSpeed*rl.GetFrameTime(), 0)
		}

		if rl.IsKeyDown(rl.KeyQ) {
			rl.CameraMoveUp(&camera, -cameraMoveSpeed*rl.GetFrameTime())
		}
		if rl.IsKeyDown(rl.KeyE) {
			rl.CameraMoveUp(&camera, cameraMoveSpeed*rl.GetFrameTime())
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
			camera.Position = rl.NewVector3(0, 4.8, 3.8)
			camera.Target = rl.NewVector3(0, 2.2, 0)
		}
		imgui.End()

		imgui.SetNextWindowPosV(imgui.NewVec2(width-12, height-12), imgui.CondAlways, imgui.NewVec2(1, 1))
		imgui.SetNextWindowSizeV(imgui.NewVec2(300, 0), imgui.CondOnce)
		imgui.BeginV("Camera", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoMove)
		imgui.SliderFloat("Move Speed", &cameraMoveSpeed, 1, 1000)
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

		if container != nil {
			for i, entry := range container.SCDContainer.Header.SceneObjects {
				rl.PushMatrix()
				rl.Translatef(entry.Position.X, entry.Position.Y, entry.Position.Z)

				rl.DrawCubeV(
					rl.Vector3Zero(),
					rl.Vector3One().Scale(20),
					rl.Yellow,
				)

				mdbContainter := container.SCDContainer.MDBContainers[i]

				for _, vb := range mdbContainter.VertexBuffers {
					for _, pos := range vb.ContainerPositions.Entries {
						rl.DrawCubeV(
							rl.NewVector3(pos.X, pos.Y, pos.Z),
							rl.Vector3One().Scale(5),
							rl.Purple,
						)
					}
				}

				rl.PopMatrix()
			}
		}

		rl.EndMode3D()

		rlig.Render()
		rl.EndDrawing()
	}

	return nil
}
