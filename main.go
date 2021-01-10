package main

import (
	"log"
	"time"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	target                rl.Vector2
	keysExemptFromRelease = []rl.Key{
		rl.KeyLeftControl,
	}
)

type Pixel struct {
	Color rl.Color
}

type Keymap map[string][][]rl.Key

type Action struct {
	// Tool is the tool used for the action (GetName is used in history panel)
	Tool Tool

	LastPixelState map[IntVec2]Pixel
}

// Tool implementations should call rl.DrawPixel or other operations, there are
// no canvas middleware
type Tool interface {
	// Used by every tool
	MouseDown(x, y int) // Called each frame the mouse is down
	MouseUp(x, y int)   // Called once, when the mouse button is released

	// Used by drawing tools
	SetColor(rl.Color)
	GetColor() rl.Color

	// Takes the current mouse position. Called every frame the tool is
	// selected. Draw calls are drawn on the preview layer.
	DrawPreview(x, y int)
}

// Layer has a Canvas and initialFill keeps track of if it's been filled on
// creation
type Layer struct {
	Canvas      rl.RenderTexture2D
	initialFill bool

	// PixelData is the "raw" pixels map
	PixelData map[IntVec2]Pixel
}

func NewLayer(width, height int, shouldFill bool) *Layer {
	return &Layer{
		Canvas:      rl.LoadRenderTexture(width, height),
		initialFill: shouldFill,
		PixelData:   make(map[IntVec2]Pixel),
	}
}

// File handles all of the actions for a file. Multiple files can be created
// and handled at the same time
type File struct {
	Camera rl.Camera2D

	Layers       []*Layer // The last one is for tool previews
	CurrentLayer int

	LeftTool            Tool
	RightTool           Tool
	HasDoneMouseUpLeft  bool
	HasDoneMouseUpRight bool

	KeyRepeat      time.Duration
	keyRepeatTimer float32
	keyMovable     bool
	lastKey        []rl.Key
	// current keys down, used for combinations
	keysDown map[rl.Key]bool
	// keys which need to be released before they can be used again
	keysAwaitingRelease map[rl.Key]bool

	// Probably a cleaner way to handle mouse relational movement...
	mouseX, mouseY, mouseLastX, mouseLastY int

	CanvasWidth, CanvasHeight, TileWidth, TileHeight int

	Keymap Keymap
}

// Update checks for input and uses the current tool to draw to the current
// layer
func (f *File) Update() {
	layer := f.Layers[f.CurrentLayer]

	// Update camera
	// TODO zoom at cursor location, not target/offset
	f.Camera.Zoom += float32(rl.GetMouseWheelMove()) * 0.1 * f.Camera.Zoom

	f.Camera.Offset.X = float32(rl.GetScreenWidth()) / 2
	f.Camera.Offset.Y = float32(rl.GetScreenHeight()) / 2
	// Move target
	f.mouseX = rl.GetMouseX()
	f.mouseY = rl.GetMouseY()
	if rl.IsMouseButtonDown(rl.MouseMiddleButton) {
		target.X += float32(f.mouseLastX-f.mouseX) / f.Camera.Zoom
		target.Y += float32(f.mouseLastY-f.mouseY) / f.Camera.Zoom
	}
	f.mouseLastX = f.mouseX
	f.mouseLastY = f.mouseY
	f.Camera.Target = target

	// Draw
	rl.BeginTextureMode(layer.Canvas)
	if !layer.initialFill {
		rl.ClearBackground(rl.DarkGray)
		layer.initialFill = true
	}

	// Handle keyboard actions
	for key := range f.keysAwaitingRelease {
		if !rl.IsKeyDown(key) {
			delete(f.keysAwaitingRelease, key)
		}
	}

	checkDown := func(keySlices [][]rl.Key) bool {
		for _, keySlice := range keySlices {
			for _, key := range keySlice {
				isDown := rl.IsKeyDown(key)
				f.keysDown[key] = isDown
				needsRelease, ok := f.keysAwaitingRelease[key]
				if !isDown || (ok && needsRelease) {
					return false
				}
			}
		}
		return true
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
	switch {
	case checkDown(f.Keymap["undo"]) && setAwaitingRelease(f.Keymap["undo"]):
		f.Undo()
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
	checkDownAddStack(f.Keymap["toolRight"])
	checkDownAddStack(f.Keymap["toolLeft"])
	checkDownAddStack(f.Keymap["toolDown"])
	checkDownAddStack(f.Keymap["toolUp"])

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
			// TODO move amount based on zoom
			switch {
			case matches(last, f.Keymap["toolRight"]):
				rl.SetMousePosition(x+moveAmount, y)
			case matches(last, f.Keymap["toolLeft"]):
				rl.SetMousePosition(x-moveAmount, y)
			case matches(last, f.Keymap["toolDown"]):
				rl.SetMousePosition(x, y+moveAmount)
			case matches(last, f.Keymap["toolUp"]):
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

	cursor := rl.GetScreenToWorld2D(rl.GetMousePosition(), f.Camera)
	cursor = cursor.Add(rl.NewVector2(float32(layer.Canvas.Texture.Width)/2, float32(layer.Canvas.Texture.Height)/2))
	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		f.HasDoneMouseUpLeft = false
		f.LeftTool.MouseDown(int(cursor.X), int(cursor.Y))
	} else {
		if f.HasDoneMouseUpLeft == false {
			f.HasDoneMouseUpLeft = true
			f.LeftTool.MouseUp(int(cursor.X), int(cursor.Y))
		}
	}
	if rl.IsMouseButtonDown(rl.MouseRightButton) {
		f.HasDoneMouseUpRight = false
		f.RightTool.MouseDown(int(cursor.X), int(cursor.Y))
	} else {
		if f.HasDoneMouseUpRight == false {
			f.HasDoneMouseUpRight = true
			f.RightTool.MouseUp(int(cursor.X), int(cursor.Y))
		}
	}
	rl.EndTextureMode()

	rl.BeginTextureMode(f.Layers[len(f.Layers)-1].Canvas)
	// LeftTool draws last as it's more important
	f.RightTool.DrawPreview(int(cursor.X), int(cursor.Y))
	f.LeftTool.DrawPreview(int(cursor.X), int(cursor.Y))
	rl.EndTextureMode()
}

// Draw is used to draw all of the layers
func (f *File) Draw() {
	rl.BeginMode2D(f.Camera)
	for _, layer := range f.Layers {
		rl.DrawTextureRec(layer.Canvas.Texture,
			rl.NewRectangle(0, 0, float32(layer.Canvas.Texture.Width), -float32(layer.Canvas.Texture.Height)),
			rl.NewVector2(-float32(layer.Canvas.Texture.Width)/2, -float32(layer.Canvas.Texture.Height)/2),
			rl.White)
	}

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
}

// Undo an action
func (f *File) Undo() {
	log.Println("undoing")
}

// Destroy unloads each layer's canvas
func (f *File) Destroy() {
	for _, layer := range f.Layers {
		layer.Canvas.Unload()
	}
}

// NewFile is the constructor for File
func NewFile(keymap Keymap, canvasWidth, canvasHeight, tileWidth, tileHeight int) *File {
	f := &File{
		Layers: []*Layer{
			NewLayer(canvasWidth, canvasHeight, false),
			NewLayer(canvasWidth, canvasHeight, true),
		},
		HasDoneMouseUpLeft:  true,
		HasDoneMouseUpRight: true,
		KeyRepeat:           time.Second / 5,
		Keymap:              keymap,
		Camera:              rl.Camera2D{Zoom: 8.0},
		CanvasWidth:         canvasWidth,
		CanvasHeight:        canvasHeight,
		TileWidth:           tileWidth,
		TileHeight:          tileHeight,
		keysDown:            make(map[rl.Key]bool),
		keysAwaitingRelease: make(map[rl.Key]bool),
	}
	f.LeftTool = NewPixelBrushTool(rl.Red, f)
	f.RightTool = NewPixelBrushTool(rl.Green, f)
	return f
}

func main() {
	log.SetFlags(log.Lshortfile)

	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 450, "Pixel")
	rl.SetTargetFPS(120)

	keymap := Keymap{
		"toolLeft":  {{rl.KeyH}, {rl.KeyLeft}},
		"toolRight": {{rl.KeyN}, {rl.KeyRight}},
		"toolUp":    {{rl.KeyC}, {rl.KeyUp}},
		"toolDown":  {{rl.KeyT}, {rl.KeyDown}},
		"undo":      {{rl.KeyLeftControl, rl.KeyZ}},
	}

	file := NewFile(keymap, 64, 64, 8, 8)

	for !rl.WindowShouldClose() {

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// Update and draw to texture using current tool
		file.Update()

		// Draw the file.Canvas, use the camera to draw file.Canvas in the correct place
		file.Draw()

		rl.EndDrawing()
	}

	// Destroy resources
	file.Destroy()

	rl.CloseWindow()
}
