package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

func main() {
	log.SetFlags(log.Lshortfile)

	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 450, "Pixel")
	rl.SetTargetFPS(120)

	keymap := KeymapData{
		"toolLeft":  {{rl.KeyH}, {rl.KeyLeft}},
		"toolRight": {{rl.KeyN}, {rl.KeyRight}},
		"toolUp":    {{rl.KeyC}, {rl.KeyUp}},
		"toolDown":  {{rl.KeyT}, {rl.KeyDown}},
		"undo":      {{rl.KeyLeftControl, rl.KeyZ}},
		"redo":      {{rl.KeyLeftControl, rl.KeyLeftShift, rl.KeyZ}, {rl.KeyLeftControl, rl.KeyY}},
	}

	InitUI()
	file := NewFile(NewKeymap(keymap), 64, 64, 8, 8)

	for !rl.WindowShouldClose() {

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		file.Update()

		rl.EndDrawing()
	}

	// Destroy resources
	file.Destroy()

	rl.CloseWindow()
}
