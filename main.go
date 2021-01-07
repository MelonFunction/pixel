package main

import (
	"log"
	"time"

	rl "github.com/lachee/raylib-goplus/raylib"
	"github.com/scarycoffee/pixel/tools"
)

var (
	camera rl.Camera2D
	// canvas rl.RenderTexture2D
	target rl.Vector2
)

type Keymap map[string]rl.Key

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

type Layer struct {
	Canvas      rl.RenderTexture2D
	initialFill bool
}

type File struct {
	// Layers belonging to the canvas. The last one is for tool previews
	Layers       []*Layer
	CurrentLayer int

	CurrentTool    Tool
	HasDoneMouseUp bool

	KeyRepeat      time.Duration
	keyRepeatTimer float32
	keyMovable     bool
	lastKey        []rl.Key

	Zoom float32 // Camera zoom for pixel movement

	Keymap Keymap
}

// Update checks for input and uses the current tool to draw to the current
// layer
func (c *File) Update() {
	layer := c.Layers[c.CurrentLayer]

	rl.BeginTextureMode(layer.Canvas)
	if !layer.initialFill {
		rl.ClearBackground(rl.DarkGray)
		layer.initialFill = true
	}

	c.keyRepeatTimer += rl.GetFrameTime() * 1000
	if c.keyRepeatTimer > float32(c.KeyRepeat.Milliseconds()) {
		c.keyRepeatTimer = 0
		c.keyMovable = true
	}

	// Stack keys up so that if left is held, then right is held, then right
	// is released, the cursor would continue going left instead of staying
	// still
	if rl.IsKeyPressed(rl.KeyN) {
		c.keyMovable = true
		c.lastKey = append(c.lastKey, rl.KeyN)
	}
	if rl.IsKeyPressed(rl.KeyH) {
		c.keyMovable = true
		c.lastKey = append(c.lastKey, rl.KeyH)
	}
	if rl.IsKeyPressed(rl.KeyC) {
		c.keyMovable = true
		c.lastKey = append(c.lastKey, rl.KeyC)
	}
	if rl.IsKeyPressed(rl.KeyT) {
		c.keyMovable = true
		c.lastKey = append(c.lastKey, rl.KeyT)
	}

	if len(c.lastKey) > 0 && rl.IsKeyDown(c.lastKey[len(c.lastKey)-1]) {
		last := c.lastKey[len(c.lastKey)-1]
		if c.keyMovable {
			c.keyRepeatTimer = 0
			c.keyMovable = false

			moveAmount := int(c.Zoom)
			x := rl.GetMouseX()
			y := rl.GetMouseY()

			// TODO move amount based on zoom
			switch last {
			case c.Keymap["toolRight"]:
				rl.SetMousePosition(x+moveAmount, y)
			case c.Keymap["toolLeft"]:
				rl.SetMousePosition(x-moveAmount, y)
			case c.Keymap["toolDown"]:
				rl.SetMousePosition(x, y+moveAmount)
			case c.Keymap["toolUp"]:
				rl.SetMousePosition(x, y-moveAmount)
			}
		}
	} else {
		// Pop lastKey until we find a key that's still down
		if len(c.lastKey) > 0 {
			c.lastKey = c.lastKey[:len(c.lastKey)-1]
		}
		c.keyRepeatTimer = 0
		c.keyMovable = true
	}

	cursor := rl.GetScreenToWorld2D(rl.GetMousePosition(), camera)
	cursor = cursor.Add(rl.NewVector2(float32(layer.Canvas.Texture.Width)/2, float32(layer.Canvas.Texture.Height)/2))
	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		c.HasDoneMouseUp = false
		c.CurrentTool.MouseDown(int(cursor.X), int(cursor.Y))
	} else {
		if c.HasDoneMouseUp == false {
			c.HasDoneMouseUp = true
			c.CurrentTool.MouseUp(int(cursor.X), int(cursor.Y))
		}
	}
	rl.EndTextureMode()

	rl.BeginTextureMode(c.Layers[len(c.Layers)-1].Canvas)
	c.CurrentTool.DrawPreview(int(cursor.X), int(cursor.Y))
	rl.EndTextureMode()
}

// Draw is used to draw all of the layers
func (c *File) Draw() {
	for _, layer := range c.Layers {
		rl.DrawTextureRec(layer.Canvas.Texture,
			rl.NewRectangle(0, 0, float32(layer.Canvas.Texture.Width), -float32(layer.Canvas.Texture.Height)),
			rl.NewVector2(-float32(layer.Canvas.Texture.Width)/2, -float32(layer.Canvas.Texture.Height)/2),
			rl.White)
	}
}
func (c *File) Destroy() {
	for _, layer := range c.Layers {
		layer.Canvas.Unload()
	}
}
func NewFile(keymap Keymap) *File {
	return &File{
		Layers: []*Layer{
			{rl.LoadRenderTexture(64, 64), false},
			{rl.LoadRenderTexture(64, 64), true},
		},
		CurrentTool:    &tools.PixelBrushTool{Color: rl.Red},
		HasDoneMouseUp: true,
		KeyRepeat:      time.Second / 5,
		Zoom:           1,
		Keymap:         keymap,
	}
}

func main() {
	log.SetFlags(log.Lshortfile)

	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 450, "Pixel")
	rl.SetTargetFPS(120)

	keymap := Keymap{
		"toolLeft":  rl.KeyH,
		"toolRight": rl.KeyN,
		"toolUp":    rl.KeyC,
		"toolDown":  rl.KeyT,
	}

	file := NewFile(keymap)

	camera = rl.Camera2D{}
	camera.Zoom = 8.0

	var mouseX, mouseY, mouseLastX, mouseLastY int
	for !rl.WindowShouldClose() {

		// TODO zoom at cursor location, not target/offset
		camera.Zoom += float32(rl.GetMouseWheelMove()) * 0.1 * camera.Zoom
		file.Zoom = camera.Zoom

		camera.Offset.X = float32(rl.GetScreenWidth()) / 2
		camera.Offset.Y = float32(rl.GetScreenHeight()) / 2
		// Move target
		mouseX = rl.GetMouseX()
		mouseY = rl.GetMouseY()
		if rl.IsMouseButtonDown(rl.MouseMiddleButton) {
			target.X += float32(mouseLastX-mouseX) / camera.Zoom
			target.Y += float32(mouseLastY-mouseY) / camera.Zoom
		}
		mouseLastX = mouseX
		mouseLastY = mouseY
		camera.Target = target

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// Update and draw to texture using current tool
		file.Update()

		// Draw the file.Canvas, use the camera to draw file.Canvas in the correct place
		rl.BeginMode2D(camera)
		file.Draw()

		rl.EndMode2D()
		rl.EndDrawing()
	}

	// Destroy resources
	file.Destroy()

	rl.CloseWindow()
}
