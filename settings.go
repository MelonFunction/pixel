package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

// SettingsData is the settings object which is read from settings.json
type SettingsData struct {
	KeymapData  KeymapData  `binding:"required"`
	PaletteData PaletteData `binding:"required"`
}

// KeymapData stores the action name as the key and a 2d slice of the keys
type KeymapData map[string][][]rl.Key

// Keymap stores the command+actions in Map and the the ordered keys in Keys
// This is usable by system_controls.go
type Keymap struct {
	Keys []string
	Data KeymapData
}

// NewKeymap returns a new Keymap
// It also sorts the keys to avoid conflicts between bindings as ctrl+z will
// fire before ctrl+shift+z if it is called first. Longer similar bindings will
// be before shorter similar ones in the list to ensure that they will always be
// called before a shorter binding is
func NewKeymap(data KeymapData) Keymap {
	keys := make([]string, 0, 0)

	for name, outer := range data {

		var longestInner []rl.Key
		for _, inner := range outer {
			if len(inner) > len(longestInner) {
				longestInner = inner
			}
		}
		didInsert := false
		for i, k := range keys {
			for _, inner := range data[k] {
				if len(longestInner) > len(inner) && !didInsert {
					didInsert = true
					keys = append(keys[:i], append([]string{name}, keys[i:]...)...)
				}
			}
		}

		if !didInsert {
			keys = append(keys, name)
		}
	}

	return Keymap{
		Keys: keys,
		Data: data,
	}
}

// PaletteData contains all of the Palettes
type PaletteData []Palette

// Palette is a list of the colors found in the palette
type Palette struct {
	Name string
	// Usable rl.Color data
	data []rl.Color
	// Hex which is converted to rl.Color on read, overwrites everything in data
	Strings []string
}

var (
	// Settings is the global settings object
	Settings *SettingsData

	defaultKeymap = KeymapData{
		"toggleGrid": {{rl.KeyG}},
		"showDebug":  {{rl.KeyD}},
		"resize":     {{rl.KeyLeftControl, rl.KeyR}},

		"pixelBrush": {{rl.KeyB}},
		"eraser":     {{rl.KeyE}},
		"fill":       {{rl.KeyF}},
		"picker":     {{rl.KeyM}},
		"selector":   {{rl.KeyS}},

		"flipHorizontal": {{rl.KeyZ}},
		"flipVertical":   {{rl.KeyV}},

		"layerUp":   {{rl.KeyLeftShift, rl.KeyUp}},
		"layerDown": {{rl.KeyLeftShift, rl.KeyDown}},

		"toolLeft":  {{rl.KeyH}, {rl.KeyLeft}},
		"toolRight": {{rl.KeyN}, {rl.KeyRight}},
		"toolUp":    {{rl.KeyC}, {rl.KeyUp}},
		"toolDown":  {{rl.KeyT}, {rl.KeyDown}},

		"cancel": {{rl.KeyEscape}},
		"copy":   {{rl.KeyLeftControl, rl.KeyC}},
		"paste":  {{rl.KeyLeftControl, rl.KeyV}},
		"delete": {{rl.KeyDelete}},

		"open":   {{rl.KeyLeftControl, rl.KeyO}},
		"save":   {{rl.KeyLeftControl, rl.KeyS}},
		"export": {{rl.KeyLeftControl, rl.KeyE}},
		"undo":   {{rl.KeyLeftControl, rl.KeyZ}},
		"redo":   {{rl.KeyLeftControl, rl.KeyLeftShift, rl.KeyZ}, {rl.KeyLeftControl, rl.KeyY}},
	}

	defaultPalettes = PaletteData{
		{
			Name: "Default",
			data: []rl.Color{
				// rl.Red,
				// rl.Blue,
				// rl.Green,
				// rl.Pink,
				// rl.Orange,
				// rl.Purple,
				// rl.Aqua,
			},
			Strings: []string{
				"ff0000ff",
				"00ff00ff",
				"0000ffff",
			},
		},
	}
)

// SaveSettings writes the settings object into settings.json
func SaveSettings() error {
	// Save each color as a hex
	for pi, palette := range Settings.PaletteData {
		palette.Strings = make([]string, 0)
		for _, color := range palette.data {
			palette.Strings = append(palette.Strings, ColorToHex(color))
		}
		Settings.PaletteData[pi] = palette
	}

	j, err := json.MarshalIndent(Settings, "", "  ")
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

// LoadSettings loads the settings from the settings.json file, validates it
// or creates a new settings.json file if one doesn't exist already
func LoadSettings() error {
	Settings = &SettingsData{}
	data, err := ioutil.ReadFile("./settings.json")
	// settings.json not found or is empty
	if err != nil {
		// Make a default settings file using the default data
		Settings.KeymapData = defaultKeymap
		Settings.PaletteData = defaultPalettes
		if err := SaveSettings(); err != nil {
			return err
		}

		log.Println("👍 settings.json was missing, defaults written to file!")

	} else {
		// Setting file found
		if err := json.Unmarshal(data, Settings); err != nil {
			log.Println(err)
		}

		// Create the defaults and add them to the settings struct
		// If there is an error with unmarshalling, everything below will be added
		if keymap := Settings.KeymapData; keymap == nil {
			// TODO validate all fields
			Settings.KeymapData = defaultKeymap
			log.Println("⌨️ Keymap was missing from settings, default added")
		}
		if palettes := Settings.PaletteData; palettes == nil {
			Settings.PaletteData = defaultPalettes
			log.Println("🎨 Palettes were missing from settings, default added")
		}
		// Convert hex to rl.Color
		for pi, palette := range Settings.PaletteData {
			palette.data = make([]rl.Color, 0)
			for _, hex := range palette.Strings {
				if color, err := HexToColor(hex); err == nil {
					palette.data = append(palette.data, color)
				}
			}
			Settings.PaletteData[pi] = palette
		}

		if err := SaveSettings(); err != nil {
			return err
		}

		log.Println("👍 Loaded settings successfully!")
	}

	return nil
}
