package main

import (
	"log"
	"time"

	rl "github.com/lachee/raylib-goplus/raylib"
)

// Tool is the interface for Tool elements
type Tool interface {
	// Used by every tool
	MouseDown(x, y int) // Called each frame the mouse is down
	MouseUp(x, y int)   // Called once, when the mouse button is released

	// Used by drawing tools
	SetColor(rl.Color)
	GetColor() rl.Color

	String() string

	// Takes the current mouse position. Called every frame the tool is
	// selected. Draw calls are drawn on the preview layer.
	DrawPreview(x, y int)
}

// PixelStateData stores what the state was previously and currently
// Prev is used by undo and Current is used by redo
type PixelStateData struct {
	Prev, Current rl.Color
}

// LayerStateData stores which layer the actions happened on
type LayerStateData struct {
	Prev, Current int
}

// HistoryAction stores information about the action
type HistoryAction struct {
	// Tool is the tool used for the action (GetName is used in history panel)
	Tool Tool

	PixelState map[IntVec2]PixelStateData
	LayerState LayerStateData
}

// KeymapData stores the action name as the key and a 2d slice of the keys
type KeymapData map[string][][]rl.Key

// Keymap stores the command+actions in Map and the the ordered keys in Keys
type Keymap struct {
	Keys []string
	Data KeymapData
}

// NewKeymap returns a new Keymap
// It also sorts the keys to avoid conflicts between bindings as ctrl+z will
// fire before ctrl+shift+z if it is called first. Longer similar bindings will
// be before shorter similar ones in the list
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

// Static vars for file
var (
	keysExemptFromRelease = []rl.Key{
		rl.KeyLeftControl,
		rl.KeyLeftShift,
		rl.KeyRightControl,
		rl.KeyRightShift,
		rl.KeyLeftAlt,
		rl.KeyRightAlt,
	}
)

// DrawPixel draws a pixel. It records actions into history.
func (f *File) DrawPixel(x, y int, color rl.Color, saveToHistory bool) {
	rl.DrawPixel(x, y, color)

	// Set the pixel data in the current layer
	layer := f.GetCurrentLayer()
	if x >= 0 && y >= 0 && x < f.CanvasWidth && y < f.CanvasHeight {
		// Add old color to history
		if saveToHistory {
			oldColor, ok := layer.PixelData[IntVec2{x, y}]
			if ok {

				// Prevent overwriting the old color with the new color since this
				// function is called every frame
				// Always draws to the last element of f.History since the
				// offset is removed automatically
				if oldColor != color {
					ps := f.History[len(f.History)-1].PixelState[IntVec2{x, y}]
					ps.Current = color
					ps.Prev = oldColor
					f.History[len(f.History)-1].PixelState[IntVec2{x, y}] = ps
				}
			}
		}

		// Change pixel data to the new color
		layer.PixelData[IntVec2{x, y}] = color
	}
}

// ClearBackground fills out the initial PixelData. Probably shouldn't be
// called every frame.
func (f *File) ClearBackground(color rl.Color) {
	rl.ClearBackground(color)

	layer := f.GetCurrentLayer()
	for x := 0; x < f.CanvasWidth; x++ {
		for y := 0; y < f.CanvasHeight; y++ {
			layer.PixelData[IntVec2{x, y}] = color
		}
	}
}

// File handles all of the actions for a file. Multiple files can be created
// and handled at the same time
type File struct {
	Camera rl.Camera2D
	target rl.Vector2

	Layers       []*Layer // The last one is for tool previews
	CurrentLayer int

	History           []HistoryAction
	HistoryMaxActions int
	historyOffset     int // How many undos have been made

	Tools               []Tool // TODO Will be used by a Tool selector UI
	LeftTool            Tool
	RightTool           Tool
	HasDoneMouseUpLeft  bool
	HasDoneMouseUpRight bool

	UI map[string]*Entity

	Keymap              Keymap
	KeyRepeat           time.Duration
	keyRepeatTimer      float32
	keyMovable          bool
	lastKey             []rl.Key
	keysDown            map[rl.Key]bool // current keys down, used for combinations
	keysAwaitingRelease map[rl.Key]bool // keys which need to be released before they can be used again

	// Used for relational mouse movement
	mouseX, mouseY, mouseLastX, mouseLastY int

	CanvasWidth, CanvasHeight, TileWidth, TileHeight int
}

// NewFile returns a pointer to a new File
func NewFile(keymap Keymap, canvasWidth, canvasHeight, tileWidth, tileHeight int) *File {
	f := &File{
		Camera: rl.Camera2D{Zoom: 8.0},

		Layers: []*Layer{
			NewLayer(canvasWidth, canvasHeight, "background", true),
			NewLayer(canvasWidth, canvasHeight, "layer 1", false),
			NewLayer(canvasWidth, canvasHeight, "layer 2", false),
			NewLayer(canvasWidth, canvasHeight, "hidden", false),
		},
		History:           make([]HistoryAction, 0, 5),
		HistoryMaxActions: 5, // TODO get from config

		HasDoneMouseUpLeft:  true,
		HasDoneMouseUpRight: true,

		KeyRepeat:           time.Second / 5,
		Keymap:              keymap,
		keysDown:            make(map[rl.Key]bool),
		keysAwaitingRelease: make(map[rl.Key]bool),

		CanvasWidth:  canvasWidth,
		CanvasHeight: canvasHeight,
		TileWidth:    tileWidth,
		TileHeight:   tileHeight,
	}
	f.LeftTool = NewPixelBrushTool(rl.Red, f, "Pixel Brush L")
	f.RightTool = NewPixelBrushTool(rl.Green, f, "Pixel Brush R")

	f.UI = map[string]*Entity{
		"layers": NewLayersUI(rl.NewRectangle(float32(rl.GetScreenWidth()-256), float32(rl.GetScreenHeight()-400), 256, 400), f),
	}

	f.Camera.Offset.X = float32(rl.GetScreenWidth()) / 2
	f.Camera.Offset.Y = float32(rl.GetScreenHeight()) / 2

	return f
}

// SetCurrentLayer sets the current layer
func (f *File) SetCurrentLayer(index int) {
	if len(f.History) > 0 && f.historyOffset == 0 {
		// Only record layer switches if we don't have any undos
		f.History[len(f.History)-1].LayerState.Prev = f.CurrentLayer
		f.History[len(f.History)-1].LayerState.Prev = index
	}

	// for i, h := range f.History {
	// 	log.Println(i, h.LayerState, len(h.PixelState))
	// }
	// log.Println()
	f.CurrentLayer = index
}

// GetCurrentLayer reutrns the current layer
func (f *File) GetCurrentLayer() *Layer {
	return f.Layers[f.CurrentLayer]
}

func (f *File) AddNewLayer() {
	newLayer := NewLayer(f.CanvasWidth, f.CanvasHeight, "new layer", false)
	f.Layers = append(f.Layers[:len(f.Layers)-1], newLayer, f.Layers[len(f.Layers)-1])
	f.SetCurrentLayer(len(f.Layers) - 2) // -2 bc temp layer is excluded
}

// Update checks for input and uses the current tool to draw to the current
// layer
func (f *File) Update() {
	UpdateUI()

	layer := f.GetCurrentLayer()

	f.mouseX = rl.GetMouseX()
	f.mouseY = rl.GetMouseY()

	// Scroll towards the cursor's location
	if !UIHasControl {
		scrollAmount := rl.GetMouseWheelMove()
		if scrollAmount != 0 {
			// TODO scroll scalar in config (0.1)
			f.target.X += ((float32(f.mouseX) - float32(rl.GetScreenWidth())/2) / (f.Camera.Zoom * 10)) * float32(scrollAmount)
			f.target.Y += ((float32(f.mouseY) - float32(rl.GetScreenHeight())/2) / (f.Camera.Zoom * 10)) * float32(scrollAmount)
			f.Camera.Target = f.target
			f.Camera.Zoom += float32(scrollAmount) * 0.1 * f.Camera.Zoom
		}
	}

	// Move target
	if rl.IsWindowResized() {
		f.Camera.Offset.X = float32(rl.GetScreenWidth()) / 2
		f.Camera.Offset.Y = float32(rl.GetScreenHeight()) / 2

		// Should probably make something that snaps components to others or
		// to the window edge but that's a problem for another day (TODO)
		for name, entity := range f.UI {
			if res, err := scene.QueryID(entity.ID); err == nil {
				moveable := res.Components[entity.Scene.ComponentsMap["moveable"]].(*Moveable)

				switch name {
				case "layers":
					moveable.Bounds.X = float32(rl.GetScreenWidth()) - moveable.Bounds.Width
					moveable.Bounds.Y = float32(rl.GetScreenHeight()) - moveable.Bounds.Height
					entity.FlowChildren()
				}
			}

		}

	}

	if rl.IsMouseButtonDown(rl.MouseMiddleButton) {
		f.target.X += float32(f.mouseLastX-f.mouseX) / f.Camera.Zoom
		f.target.Y += float32(f.mouseLastY-f.mouseY) / f.Camera.Zoom
	}
	f.mouseLastX = f.mouseX
	f.mouseLastY = f.mouseY
	f.Camera.Target = f.target

	// Draw
	rl.BeginTextureMode(layer.Canvas)
	if !layer.hasInitialFill {
		f.ClearBackground(rl.DarkGray)
		layer.hasInitialFill = true
	}

	// Handle keyboard actions
	for key := range f.keysAwaitingRelease {
		if !rl.IsKeyDown(key) {
			delete(f.keysAwaitingRelease, key)
		}
	}

	checkDown := func(keySlices [][]rl.Key) bool {
		allDown := true
		for _, keySlice := range keySlices {
			// Reset for each combination for the binding
			allDown = true
			for _, key := range keySlice {
				isDown := rl.IsKeyDown(key)
				f.keysDown[key] = isDown
				needsRelease, ok := f.keysAwaitingRelease[key]
				if !isDown || (ok && needsRelease) {
					allDown = false
				}
			}
			if allDown {
				return allDown
			}
		}
		return allDown
	}
	setAwaitingRelease := func(keySlices [][]rl.Key) bool {
		for _, keySlice := range keySlices {
			for _, key := range keySlice {
				found := false
				for _, k := range keysExemptFromRelease {
					if k == key {
						found = true
					}
				}
				if !found {
					f.keysAwaitingRelease[key] = true
				}
			}
		}
		return true
	}

	// If checkDown is true, then execute setAwaitingRelease (return isn't important)
	for _, key := range f.Keymap.Keys {
		if checkDown(f.Keymap.Data[key]) {
			setAwaitingRelease(f.Keymap.Data[key])

			switch key {
			case "undo":
				f.Undo()
			case "redo":
				f.Redo()
			}

			break
		}
	}

	f.keyRepeatTimer += rl.GetFrameTime() * 1000
	if f.keyRepeatTimer > float32(f.KeyRepeat.Milliseconds()) {
		f.keyRepeatTimer = 0
		f.keyMovable = true
	}
	// Stack keys up so that if left is held, then right is held, then right
	// is released, the cursor would continue going left instead of staying
	// still
	checkDownAddStack := func(keySlices [][]rl.Key) {
		for _, keySlice := range keySlices {
			for _, key := range keySlice {
				if rl.IsKeyPressed(key) {
					f.keyMovable = true
					f.lastKey = append(f.lastKey, key)
				}
			}
		}
	}
	checkDownAddStack(f.Keymap.Data["toolRight"])
	checkDownAddStack(f.Keymap.Data["toolLeft"])
	checkDownAddStack(f.Keymap.Data["toolDown"])
	checkDownAddStack(f.Keymap.Data["toolUp"])

	if len(f.lastKey) > 0 && rl.IsKeyDown(f.lastKey[len(f.lastKey)-1]) {
		last := f.lastKey[len(f.lastKey)-1]
		if f.keyMovable {
			f.keyRepeatTimer = 0
			f.keyMovable = false

			moveAmount := int(f.Camera.Zoom)
			x := rl.GetMouseX()
			y := rl.GetMouseY()

			matches := func(match rl.Key, keySlices [][]rl.Key) bool {
				for _, keySlice := range keySlices {
					for _, key := range keySlice {
						if key == match {
							return true
						}
					}
				}
				return false
			}
			switch {
			case matches(last, f.Keymap.Data["toolRight"]):
				rl.SetMousePosition(x+moveAmount, y)
			case matches(last, f.Keymap.Data["toolLeft"]):
				rl.SetMousePosition(x-moveAmount, y)
			case matches(last, f.Keymap.Data["toolDown"]):
				rl.SetMousePosition(x, y+moveAmount)
			case matches(last, f.Keymap.Data["toolUp"]):
				rl.SetMousePosition(x, y-moveAmount)
			}
		}
	} else {
		// Pop lastKey until we find a key that's still down
		if len(f.lastKey) > 0 {
			f.lastKey = f.lastKey[:len(f.lastKey)-1]
		}
		f.keyRepeatTimer = 0
		f.keyMovable = true
	}

	appendHistory := func(action HistoryAction) {
		// Clear everything past the offset if a change has been made after undoing
		f.History = f.History[0 : len(f.History)-f.historyOffset]
		f.historyOffset = 0

		if len(f.History) >= f.HistoryMaxActions {
			f.History = append(f.History[len(f.History)-f.HistoryMaxActions+1:f.HistoryMaxActions], action)
		} else {
			f.History = append(f.History, action)
		}
	}

	cursor := rl.GetScreenToWorld2D(rl.GetMousePosition(), f.Camera)
	cursor = cursor.Add(rl.NewVector2(float32(layer.Canvas.Texture.Width)/2, float32(layer.Canvas.Texture.Height)/2))
	if !UIHasControl {
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			// Fires once
			if f.HasDoneMouseUpLeft {
				// Create new history action
				appendHistory(HistoryAction{f.LeftTool, make(map[IntVec2]PixelStateData), LayerStateData{f.CurrentLayer, f.CurrentLayer}})
			}
			f.HasDoneMouseUpLeft = false

			// Repeated action
			f.LeftTool.MouseDown(int(cursor.X), int(cursor.Y))
		} else {
			// Always fires once
			if f.HasDoneMouseUpLeft == false {
				f.HasDoneMouseUpLeft = true
				f.LeftTool.MouseUp(int(cursor.X), int(cursor.Y))
			}
		}

		if rl.IsMouseButtonDown(rl.MouseRightButton) {
			if f.HasDoneMouseUpRight {
				appendHistory(HistoryAction{f.RightTool, make(map[IntVec2]PixelStateData), LayerStateData{f.CurrentLayer, f.CurrentLayer}})
			}
			f.HasDoneMouseUpRight = false
			f.RightTool.MouseDown(int(cursor.X), int(cursor.Y))
		} else {
			if f.HasDoneMouseUpRight == false {
				f.HasDoneMouseUpRight = true
				f.RightTool.MouseUp(int(cursor.X), int(cursor.Y))
			}
		}
	}
	rl.EndTextureMode()

	rl.BeginTextureMode(f.Layers[len(f.Layers)-1].Canvas)
	// LeftTool draws last as it's more important
	f.RightTool.DrawPreview(int(cursor.X), int(cursor.Y))
	f.LeftTool.DrawPreview(int(cursor.X), int(cursor.Y))
	rl.EndTextureMode()

	rl.BeginMode2D(f.Camera)
	for _, layer := range f.Layers {
		if !layer.Hidden {
			rl.DrawTextureRec(layer.Canvas.Texture,
				rl.NewRectangle(0, 0, float32(layer.Canvas.Texture.Width), -float32(layer.Canvas.Texture.Height)),
				rl.NewVector2(-float32(layer.Canvas.Texture.Width)/2, -float32(layer.Canvas.Texture.Height)/2),
				rl.White)
		}
		// log.Println(layer.Name, len(layer.PixelData))
	}

	// TODO use a high resolution texture to draw grids, then we won't need to draw lines each draw call
	for x := 0; x <= f.CanvasWidth; x += f.TileWidth {
		rl.DrawLine(
			-f.CanvasWidth/2+x,
			-f.CanvasHeight/2,
			-f.CanvasWidth/2+x,
			f.CanvasHeight/2,
			rl.White)
	}
	for y := 0; y <= f.CanvasHeight; y += f.TileHeight {
		rl.DrawLine(
			-f.CanvasWidth/2,
			-f.CanvasHeight/2+y,
			f.CanvasWidth/2,
			-f.CanvasHeight/2+y,
			rl.White)
	}
	rl.EndMode2D()

	DrawUI()
}

// Undo an action
func (f *File) Undo() {
	// TODO make layer history actually work
	if f.historyOffset < len(f.History) {
		f.historyOffset++
		log.Println(f.History[len(f.History)-f.historyOffset].LayerState)
		f.SetCurrentLayer(f.History[len(f.History)-f.historyOffset].LayerState.Prev)
		for pos, psd := range f.History[len(f.History)-f.historyOffset].PixelState {
			f.DrawPixel(pos.X, pos.Y, psd.Prev, false)
		}
	}
}

// Redo an action
func (f *File) Redo() {
	if f.historyOffset > 0 {
		for pos, psd := range f.History[len(f.History)-f.historyOffset].PixelState {
			// f.SetCurrentLayer(f.History[len(f.History)-f.historyOffset].LayerState.Current)
			f.DrawPixel(pos.X, pos.Y, psd.Current, false)
		}
		f.historyOffset--
	}
}

// Destroy unloads each layer's canvas
func (f *File) Destroy() {
	for _, layer := range f.Layers {
		layer.Canvas.Unload()
	}
	DestroyUI()
}
