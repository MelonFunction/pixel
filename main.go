package main

import (
	"log"
	"os"
	"path"

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
	rl.InitWindow(1920*0.75, 1080*0.75, "Pixel")
	rl.SetWindowPosition(1100, 500)
	rl.SetTargetFPS(60)
	rl.SetExitKey(0)

	// Make the files
	Files = []*File{}

	CurrentFile = NewFile(64, 64, 8, 8)
	Files = append(Files, CurrentFile)

	// Load the settings
	err := LoadSettings()
	if err != nil {
		log.Println(err)
	}

	InitUI(NewKeymap(Settings.KeymapData))

	if len(os.Args) > 1 {
		// delete starting/empty file
		Files = []*File{}

		for _, argPath := range os.Args[1:] {
			// Default path
			pathDir, err := os.Getwd()
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(pathDir)
			pathDir = path.Join(pathDir, argPath)
			log.Println(pathDir)

			newFile := Open(pathDir)
			Files = append(Files, newFile)
		}
	}

	// show filename(s) in tab
	EditorsUIRebuild()

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
