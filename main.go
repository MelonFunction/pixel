package main

import (
	"log"
	"time"

	rl "github.com/lachee/raylib-goplus/raylib"
	"github.com/scarycoffee/pixel/tools"
)

var (
	target rl.Vector2
)

type Keymap map[string][]rl.Key

// Tool implementations should call rl.DrawPixel or other operations, there are
// no canvas middleware
// All tools need MouseDown and MouseUp, but SetColor and GetColor are for
// the majority of functions as they will need it
type Tool interface {
	MouseDown(x, y int) // Called each frame the mouse is down
	MouseUp(x, y int)   // Called once, when the mouse button is released
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
}

// File handles all of the actions for a file. Multiple files can be created
// and handled at the same time
type File struct {
	Camera rl.Camera2D

	Layers       []*Layer // The last one is for tool previews
	CurrentLayer int

	CurrentTool    Tool
	HasDoneMouseUp bool

	KeyRepeat      time.Duration
	keyRepeatTimer float32
	keyMovable     bool
	lastKey        []rl.Key

	// Probably a cleaner way to handle mouse relational movement...
	mouseX, mouseY, mouseLastX, mouseLastY int

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

	f.keyRepeatTimer += rl.GetFrameTime() * 1000
	if f.keyRepeatTimer > float32(f.KeyRepeat.Milliseconds()) {
		f.keyRepeatTimer = 0
		f.keyMovable = true
	}

	// Stack keys up so that if left is held, then right is held, then right
	// is released, the cursor would continue going left instead of staying
	// still
	checkDown := func(keys []rl.Key) {
		for _, key := range keys {
			if rl.IsKeyPressed(key) {
				f.keyMovable = true
				f.lastKey = append(f.lastKey, key)
			}
		}
	}
	checkDown(f.Keymap["toolRight"])
	checkDown(f.Keymap["toolLeft"])
	checkDown(f.Keymap["toolDown"])
	checkDown(f.Keymap["toolUp"])

	if len(f.lastKey) > 0 && rl.IsKeyDown(f.lastKey[len(f.lastKey)-1]) {
		last := f.lastKey[len(f.lastKey)-1]
		if f.keyMovable {
			f.keyRepeatTimer = 0
			f.keyMovable = false

			moveAmount := int(f.Camera.Zoom)
			x := rl.GetMouseX()
			y := rl.GetMouseY()

			matches := func(match rl.Key, keys []rl.Key) bool {
				for _, key := range keys {
					if key == match {
						return true
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
		f.HasDoneMouseUp = false
		f.CurrentTool.MouseDown(int(cursor.X), int(cursor.Y))
	} else {
		if f.HasDoneMouseUp == false {
			f.HasDoneMouseUp = true
			f.CurrentTool.MouseUp(int(cursor.X), int(cursor.Y))
		}
	}
	rl.EndTextureMode()

	rl.BeginTextureMode(f.Layers[len(f.Layers)-1].Canvas)
	f.CurrentTool.DrawPreview(int(cursor.X), int(cursor.Y))
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
	rl.EndMode2D()
}

// Destroy unloads each layer's canvas
func (f *File) Destroy() {
	for _, layer := range f.Layers {
		layer.Canvas.Unload()
	}
}

// NewFile is the constructor for File
func NewFile(keymap Keymap) *File {
	return &File{
		Layers: []*Layer{
			{rl.LoadRenderTexture(64, 64), false},
			{rl.LoadRenderTexture(64, 64), true},
		},
		CurrentTool:    &tools.PixelBrushTool{Color: rl.Red},
		HasDoneMouseUp: true,
		KeyRepeat:      time.Second / 5,
		Keymap:         keymap,
		Camera:         rl.Camera2D{Zoom: 8.0},
	}
}

func main() {
	log.SetFlags(log.Lshortfile)

	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 450, "Pixel")
	rl.SetTargetFPS(120)

	keymap := Keymap{
		"toolLeft":  {rl.KeyH, rl.KeyLeft},
		"toolRight": {rl.KeyN, rl.KeyRight},
		"toolUp":    {rl.KeyC, rl.KeyUp},
		"toolDown":  {rl.KeyT, rl.KeyDown},
	}

	file := NewFile(keymap)

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
