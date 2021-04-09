package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	// CurrentFile is the current file being edited
	CurrentFile *File
	// Files is a slice of all the files currently loaded
	Files = make([]*File, 0, 8)
)

func main() {
	log.SetFlags(log.Lshortfile)

	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(1200, 800, "Pixel")
	rl.SetTargetFPS(60)
	// rl.SetExitKey(rl.KeyNumLock)

	keymap := KeymapData{
		"showDebug": {{rl.KeyD}},
		"layerUp":   {{rl.KeyLeftShift, rl.KeyUp}},
		"layerDown": {{rl.KeyLeftShift, rl.KeyDown}},
		"toolLeft":  {{rl.KeyH}, {rl.KeyLeft}},
		"toolRight": {{rl.KeyN}, {rl.KeyRight}},
		"toolUp":    {{rl.KeyC}, {rl.KeyUp}},
		"toolDown":  {{rl.KeyT}, {rl.KeyDown}},
		"open":      {{rl.KeyLeftControl, rl.KeyO}},
		"save":      {{rl.KeyLeftControl, rl.KeyS}},
		"export":    {{rl.KeyLeftControl, rl.KeyE}},
		"undo":      {{rl.KeyLeftControl, rl.KeyZ}},
		"redo":      {{rl.KeyLeftControl, rl.KeyLeftShift, rl.KeyZ}, {rl.KeyLeftControl, rl.KeyY}},
	}

	CurrentFile = NewFile(64, 64, 8, 8)
	Files = []*File{CurrentFile}
	InitUI(NewKeymap(keymap))

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		UpdateUI()
		DrawUI()

		rl.EndDrawing()
	}

	// Destroy resources
	for _, file := range Files {
		file.Destroy()
	}
	DestroyUI()
	UIControlSystemCmds <- "quit"

	rl.CloseWindow()
}
