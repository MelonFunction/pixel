package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

func main() {
	log.SetFlags(log.Lshortfile)

	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(1200, 800, "Pixel")
	rl.SetTargetFPS(60)

	keymap := KeymapData{
		"layerUp":   {{rl.KeyLeftShift, rl.KeyUp}},
		"layerDown": {{rl.KeyLeftShift, rl.KeyDown}},
		"toolLeft":  {{rl.KeyH}, {rl.KeyLeft}},
		"toolRight": {{rl.KeyN}, {rl.KeyRight}},
		"toolUp":    {{rl.KeyC}, {rl.KeyUp}},
		"toolDown":  {{rl.KeyT}, {rl.KeyDown}},
		"save":      {{rl.KeyLeftControl, rl.KeyS}},
		"export":    {{rl.KeyLeftControl, rl.KeyE}},
		"undo":      {{rl.KeyLeftControl, rl.KeyZ}},
		"redo":      {{rl.KeyLeftControl, rl.KeyLeftShift, rl.KeyZ}, {rl.KeyLeftControl, rl.KeyY}},
	}

	file := NewFile(64, 64, 8, 8)
	InitUI(file, NewKeymap(keymap))

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		UpdateUI()
		DrawUI()

		rl.EndDrawing()
	}

	// Destroy resources
	file.Destroy() // TODO system should handle this as there could be multiple files
	DestroyUI()

	rl.CloseWindow()
}
