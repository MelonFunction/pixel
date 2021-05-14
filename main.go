package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

// TODO make settings object and bind it to json

// Settings is the settings object which is read from settings.json
type Settings struct {
	KeymapData *KeymapData `binding:"required" json:"keymap_data"`
}

var (
	// CurrentFile is the current file being edited
	CurrentFile *File
	// Files is a slice of all the files currently loaded
	Files = make([]*File, 0, 8)
)

func SaveSettings(settings *Settings) error {
	j, err := json.Marshal(settings)
	if err != nil {
		log.Fatal(nil)
		return err
	}

	if err := ioutil.WriteFile("./settings.json", j, 0644); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func main() {
	log.SetFlags(log.Lshortfile)

	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(1200, 800, "Pixel")
	rl.SetTargetFPS(60)
	// rl.SetExitKey(rl.KeyNumLock)

	var settings *Settings

	// Create the defaults and add them to the settings struct
	defaultKeymap := &KeymapData{
		"toggleGrid": {{rl.KeyG}},
		"showDebug":  {{rl.KeyD}},
		"resize":     {{rl.KeyLeftControl, rl.KeyR}},

		"pixelBrush": {{rl.KeyB}},
		"eraser":     {{rl.KeyE}},
		"fill":       {{rl.KeyF}},
		"picker":     {{rl.KeyM}},
		"selector":   {{rl.KeyS}},

		"layerUp":   {{rl.KeyLeftShift, rl.KeyUp}},
		"layerDown": {{rl.KeyLeftShift, rl.KeyDown}},

		"toolLeft":  {{rl.KeyH}, {rl.KeyLeft}},
		"toolRight": {{rl.KeyN}, {rl.KeyRight}},
		"toolUp":    {{rl.KeyC}, {rl.KeyUp}},
		"toolDown":  {{rl.KeyT}, {rl.KeyDown}},

		"copy":  {{rl.KeyLeftControl, rl.KeyC}},
		"paste": {{rl.KeyLeftControl, rl.KeyV}},

		"open":   {{rl.KeyLeftControl, rl.KeyO}},
		"save":   {{rl.KeyLeftControl, rl.KeyS}},
		"export": {{rl.KeyLeftControl, rl.KeyE}},
		"undo":   {{rl.KeyLeftControl, rl.KeyZ}},
		"redo":   {{rl.KeyLeftControl, rl.KeyLeftShift, rl.KeyZ}, {rl.KeyLeftControl, rl.KeyY}},
	}
	settings = &Settings{}

	data, err := ioutil.ReadFile("./settings.json")
	// settings.json not found or is empty
	if err != nil {
		// Make a default settings file using the default data
		settings.KeymapData = defaultKeymap
		SaveSettings(settings)
	} else {
		// Setting file found
		if err := json.Unmarshal(data, settings); err != nil {
			log.Println(err)
		}

		// If there is an error with unmarshalling, everything below will be added
		if keymap := settings.KeymapData; keymap == nil {
			// TODO validate all fields
			settings.KeymapData = defaultKeymap
			log.Println("⌨️ Keymap was missing from settings, default added")
		}

		SaveSettings(settings)

		log.Println("👍 Loaded settings successfully!")
	}

	// json.Marshal()

	CurrentFile = NewFile(64, 64, 8, 8)
	Files = []*File{CurrentFile}
	InitUI(NewKeymap(*settings.KeymapData))

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
