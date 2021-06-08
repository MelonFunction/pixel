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

	// ShowDebug enables debug overlays when true
	ShowDebug = false
)

func main() {
	log.SetFlags(log.Lshortfile)

	SetupFiles()

	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(1200, 800, "Pixel")
	rl.SetTargetFPS(60)
	rl.SetExitKey(0)

	// Load the settings
	err := LoadSettings()
	if err != nil {
		log.Println(err)
	}

	CurrentFile = NewFile(64, 64, 8, 8)
	Files = []*File{CurrentFile}
	InitUI(NewKeymap(Settings.KeymapData))

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
