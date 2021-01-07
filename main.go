package main

import (
	"log"
	"time"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	camera rl.Camera2D
	// canvas rl.RenderTexture2D
	target rl.Vector2
)

// Tool implementations should call rl.DrawPixel or other operations, there are
// no canvas middleware
// All tools need MouseDown and MouseUp, but SetColor and GetColor are for
// the majority of functions as they will need it
type Tool interface {
	MouseDown(x, y int)
	MouseUp(x, y int)
	SetColor(rl.Color)
	GetColor() rl.Color
	DrawPrompt() // Draw a selection boundary, pixel at a certain location etc
}

type IntVec2 struct {
	X, Y int
}

type PixelTool struct {
	lastPos                IntVec2
	shouldConnectToLastPos bool
	Color                  rl.Color
}

func (t *PixelTool) MouseDown(x, y int) {
	if !t.shouldConnectToLastPos {
		t.shouldConnectToLastPos = true
		rl.DrawPixel(x, y, t.GetColor())
	} else {
		rl.DrawLine(x, y, t.lastPos.X, t.lastPos.Y, t.GetColor())
	}
	t.lastPos.X = x
	t.lastPos.Y = y
}
func (t *PixelTool) MouseUp(x, y int) {
	t.shouldConnectToLastPos = false
}
func (t *PixelTool) SetColor(color rl.Color) {
	t.Color = color
}
func (t *PixelTool) GetColor() rl.Color {
	return t.Color
}
func (t *PixelTool) DrawPrompt() {

}

type Layer struct {
	Canvas      rl.RenderTexture2D
	initialFill bool
}

type CustomCanvas struct {
	Layers       []*Layer
	CurrentLayer int

	CurrentTool Tool

	KeyRepeat      time.Duration
	keyRepeatTimer float32
	keyMovable     bool
	lastKey        []rl.Key
}

// Update checks for input and uses the current tool to draw to the current
// layer
func (c *CustomCanvas) Update() {
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

	// Queue keys up so that if left is held, then right is held, then right
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

			switch last {
			case rl.KeyN: // left
				rl.SetMousePosition(rl.GetMouseX()+10, rl.GetMouseY())
			case rl.KeyH: // right
				rl.SetMousePosition(rl.GetMouseX()-10, rl.GetMouseY())
			case rl.KeyT: // down
				rl.SetMousePosition(rl.GetMouseX(), rl.GetMouseY()+10)
			case rl.KeyC: // up
				rl.SetMousePosition(rl.GetMouseX(), rl.GetMouseY()-10)
			}
		}
	} else {
		// Stop moving
		if len(c.lastKey) > 0 {
			c.lastKey = c.lastKey[:len(c.lastKey)-1]
		}
		c.keyRepeatTimer = 0
		c.keyMovable = true
	}

	// switch {
	// case rl.IsKeyDown(rl.KeyN) || rl.IsKeyDown(rl.KeyRight):
	// 	if c.keyMovable {
	// 		c.keyRepeatTimer = 0
	// 		c.keyMovable = false
	// 		rl.SetMousePosition(rl.GetMouseX()+10, rl.GetMouseY())
	// 	}
	// case rl.IsKeyDown(rl.KeyH) || rl.IsKeyDown(rl.KeyLeft):
	// 	if c.keyMovable {
	// 		c.keyRepeatTimer = 0
	// 		c.keyMovable = false
	// 		rl.SetMousePosition(rl.GetMouseX()-10, rl.GetMouseY())
	// 	}
	// default:
	// 	c.keyRepeatTimer = 0
	// 	c.keyMovable = true
	// }

	cursor := rl.GetScreenToWorld2D(rl.GetMousePosition(), camera)
	cursor = cursor.Add(rl.NewVector2(float32(layer.Canvas.Texture.Width)/2, float32(layer.Canvas.Texture.Height)/2))
	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		c.CurrentTool.MouseDown(int(cursor.X), int(cursor.Y))
	} else {
		c.CurrentTool.MouseUp(int(cursor.X), int(cursor.Y))
	}
	rl.EndTextureMode()
}

// Draw is used to draw all of the layers
// TODO draw all of the layers
func (c *CustomCanvas) Draw() {
	for _, layer := range c.Layers {
		rl.DrawTextureRec(layer.Canvas.Texture,
			rl.NewRectangle(0, 0, float32(layer.Canvas.Texture.Width), -float32(layer.Canvas.Texture.Height)),
			rl.NewVector2(-float32(layer.Canvas.Texture.Width)/2, -float32(layer.Canvas.Texture.Height)/2),
			rl.White)
	}
}
func (c *CustomCanvas) Destroy() {
	for _, layer := range c.Layers {
		layer.Canvas.Unload()
	}
}
func NewCustomCanvas() *CustomCanvas {
	return &CustomCanvas{
		Layers: []*Layer{
			{rl.LoadRenderTexture(64, 64), false},
			{rl.LoadRenderTexture(64, 64), true}, // Scratch layer
		},
		CurrentTool: &PixelTool{Color: rl.Red},
		KeyRepeat:   time.Second / 5,
	}
}

func main() {
	log.SetFlags(log.Lshortfile)

	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 450, "Pixel")
	rl.SetTargetFPS(120)

	canvas := NewCustomCanvas()

	camera = rl.Camera2D{}
	camera.Zoom = 8.0

	var mouseX, mouseY, mouseLastX, mouseLastY int

	for !rl.WindowShouldClose() {

		// TODO zoom at cursor location, not target/offset
		camera.Zoom += float32(rl.GetMouseWheelMove()) * 0.1 * camera.Zoom

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
		canvas.Update()

		// Draw the canvas.Canvas, use the camera to draw canvas.Canvas in the correct place
		rl.BeginMode2D(camera)
		canvas.Draw()

		rl.EndMode2D()
		rl.EndDrawing()
	}

	// Destroy resources
	canvas.Destroy()

	rl.CloseWindow()
}
