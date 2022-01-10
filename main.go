package main

import (
	"log"
	"os"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	// CurrentFile is the current file being edited
	CurrentFile *File
	// Files is a slice of all the files currently loaded
	Files = make([]*File, 0, 8)

	// CopiedSelection holds the selection when File.Copy is called
	CopiedSelection map[IntVec2]rl.Color
	// CopiedSelectionPixels is a different format of the above
	CopiedSelectionPixels []rl.Color
	// IsSelectionPasted defines if the layer data should be moved or not
	IsSelectionPasted bool
	// CopiedSelectionBounds is the bounds of the copied selection
	CopiedSelectionBounds [4]int

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

	// Load the settings
	err := LoadSettings()
	if err != nil {
		log.Println(err)
	}

	CurrentFile = NewFile(64, 64, 8, 8)
	Files = append(Files, CurrentFile)

	InitUI(NewKeymap(Settings.KeymapData))

	if len(os.Args) > 1 {
		// delete starting/empty file
		Files = []*File{}

		for _, argPath := range os.Args[1:] {
			// Try using explicit/full path
			fi, err := os.Stat(argPath)
			if err == nil {
				if fi.Mode().IsRegular() {
					newFile := Open(argPath)
					Files = append(Files, newFile)
					continue
				}
			} else {
				log.Println(err)
				return
			}
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
