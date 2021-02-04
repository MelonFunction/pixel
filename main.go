package main

import (
	"log"
	"os"

	"github.com/gotk3/gotk3/gtk"
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

	go func() {
		gtk.Init(nil)

		win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
		if err != nil {
			log.Fatal("Unable to create window:", err)
		}
		win.Connect("destroy", func() {
			gtk.MainQuit()
			log.Println("destoryed")
		})

		fc, err := gtk.FileChooserNativeDialogNew(
			"Select file",
			win,
			gtk.FILE_CHOOSER_ACTION_OPEN,
			"open",
			"cancel",
		)
		acc, err := fc.GetAcceptLabel()
		log.Println(acc)

		home, err := os.UserHomeDir()
		log.Println(home)
		if err != nil {
			log.Fatal(err)
		}
		fc.SetCurrentFolder(home)

		switch fc.Run() {
		case int(gtk.RESPONSE_ACCEPT):
			log.Println("accept")
		case int(gtk.RESPONSE_CANCEL):
			log.Println("cancel")
		case int(gtk.RESPONSE_CLOSE):
			log.Println("close")
		default:
			log.Println("??")
		}

		gtk.Main()
		log.Println("done")
	}()

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
	gtk.MainQuit()

	rl.CloseWindow()
}
