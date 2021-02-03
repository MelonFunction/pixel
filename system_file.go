package main

import (
	"fmt"

	rl "github.com/lachee/raylib-goplus/raylib"
)

type UIFileSystem struct {
	BasicSystem

	UI   map[string]*Entity
	file *File

	Camera rl.Camera2D
	target rl.Vector2

	// Used for relational mouse movement
	mouseX, mouseY, mouseLastX, mouseLastY int

	cursor rl.Vector2
}

func NewUIFileSystem(file *File) *UIFileSystem {
	f := &UIFileSystem{
		Camera: rl.Camera2D{Zoom: 8.0},
		file:   file,

		UI: map[string]*Entity{
			"rgb":    NewRGBUI(rl.NewRectangle(float32(rl.GetScreenWidth()-128*1.5), float32(0), 128*1.5, 128*1.8), file),
			"layers": NewLayersUI(rl.NewRectangle(float32(rl.GetScreenWidth()-128*3), float32(rl.GetScreenHeight()-128*3), 128*3, 128*3), file),
		},
	}

	f.Camera.Offset.X = float32(rl.GetScreenWidth()) / 2
	f.Camera.Offset.Y = float32(rl.GetScreenHeight()) / 2

	return f
}

func (s *UIFileSystem) Draw() {
	layer := s.file.GetCurrentLayer()

	// Draw
	rl.BeginTextureMode(layer.Canvas)
	if !layer.hasInitialFill {
		s.file.ClearBackground(layer.InitialFillColor)
		layer.hasInitialFill = true
	}
	rl.EndTextureMode()

	// Draw temp layer
	rl.BeginTextureMode(s.file.Layers[len(s.file.Layers)-1].Canvas)
	// LeftTool draws last as it's more important
	s.file.RightTool.DrawPreview(int(s.cursor.X), int(s.cursor.Y))
	s.file.LeftTool.DrawPreview(int(s.cursor.X), int(s.cursor.Y))
	rl.EndTextureMode()

	// Draw layers
	rl.BeginMode2D(s.Camera)
	for _, layer := range s.file.Layers {
		if !layer.Hidden {
			rl.DrawTextureRec(layer.Canvas.Texture,
				rl.NewRectangle(0, 0, float32(layer.Canvas.Texture.Width), -float32(layer.Canvas.Texture.Height)),
				rl.NewVector2(-float32(layer.Canvas.Texture.Width)/2, -float32(layer.Canvas.Texture.Height)/2),
				rl.White)
		}
	}

	// Grid drawing
	// TODO use a high resolution texture to draw grids, then we won't need to draw lines each draw call
	for x := 0; x <= s.file.CanvasWidth; x += s.file.TileWidth {
		rl.DrawLine(
			-s.file.CanvasWidth/2+x,
			-s.file.CanvasHeight/2,
			-s.file.CanvasWidth/2+x,
			s.file.CanvasHeight/2,
			rl.White)
	}
	for y := 0; y <= s.file.CanvasHeight; y += s.file.TileHeight {
		rl.DrawLine(
			-s.file.CanvasWidth/2,
			-s.file.CanvasHeight/2+y,
			s.file.CanvasWidth/2,
			-s.file.CanvasHeight/2+y,
			rl.White)
	}
	rl.EndMode2D()

	// Debug text
	for y, history := range s.file.History {
		str := fmt.Sprintf("Layer: %d, Diff: %d",
			history.LayerIndex,
			len(history.PixelState))
		rl.DrawText(str, 0, 20*y, 20, rl.White)
	}

	rl.DrawText(fmt.Sprintf("Current layer: %d", s.file.CurrentLayer), 0, (s.file.HistoryMaxActions+1)*20, 20, rl.White)
	rl.DrawText(fmt.Sprintf("HistoryOffset: %d", s.file.historyOffset), 0, (s.file.HistoryMaxActions+2)*20, 20, rl.White)
}

func (s *UIFileSystem) Update(dt float32) {
	// Move target
	if rl.IsWindowResized() {
		s.Camera.Offset.X = float32(rl.GetScreenWidth()) / 2
		s.Camera.Offset.Y = float32(rl.GetScreenHeight()) / 2

		// Should probably make something that snaps components to others or
		// to the window edge but that's a problem for another day (TODO)
		for name, entity := range s.UI {
			if res, err := scene.QueryID(entity.ID); err == nil {
				moveable := res.Components[entity.Scene.ComponentsMap["moveable"]].(*Moveable)

				switch name {
				case "layers":
					moveable.Bounds.X = float32(rl.GetScreenWidth()) - moveable.Bounds.Width
					moveable.Bounds.Y = float32(rl.GetScreenHeight()) - moveable.Bounds.Height
					entity.FlowChildren()
				case "rgb":
					moveable.Bounds.X = float32(rl.GetScreenWidth()) - moveable.Bounds.Width
					entity.FlowChildren()
				}
			}

		}

	}

	layer := s.file.GetCurrentLayer()
	s.mouseX = rl.GetMouseX()
	s.mouseY = rl.GetMouseY()

	// Scroll towards the cursor's location
	if !UIHasControl {
		scrollAmount := rl.GetMouseWheelMove()
		if scrollAmount != 0 {
			// TODO scroll scalar in config (0.1)
			s.target.X += ((float32(s.mouseX) - float32(rl.GetScreenWidth())/2) / (s.Camera.Zoom * 10)) * float32(scrollAmount)
			s.target.Y += ((float32(s.mouseY) - float32(rl.GetScreenHeight())/2) / (s.Camera.Zoom * 10)) * float32(scrollAmount)
			s.Camera.Target = s.target
			s.Camera.Zoom += float32(scrollAmount) * 0.1 * s.Camera.Zoom
		}
	}

	if rl.IsMouseButtonDown(rl.MouseMiddleButton) {
		s.target.X += float32(s.mouseLastX-s.mouseX) / s.Camera.Zoom
		s.target.Y += float32(s.mouseLastY-s.mouseY) / s.Camera.Zoom
	}
	s.mouseLastX = s.mouseX
	s.mouseLastY = s.mouseY
	s.Camera.Target = s.target

	s.cursor = rl.GetScreenToWorld2D(rl.GetMousePosition(), s.Camera)
	s.cursor = s.cursor.Add(rl.NewVector2(float32(layer.Canvas.Texture.Width)/2, float32(layer.Canvas.Texture.Height)/2))
	if !UIHasControl {
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			// Fires once
			if s.file.HasDoneMouseUpLeft {
				// Create new history action
				s.file.AppendHistory(HistoryAction{make(map[IntVec2]PixelStateData), s.file.CurrentLayer})
			}
			s.file.HasDoneMouseUpLeft = false

			// Repeated action
			s.file.LeftTool.MouseDown(int(s.cursor.X), int(s.cursor.Y))
		} else {
			// Always fires once
			if s.file.HasDoneMouseUpLeft == false {
				s.file.HasDoneMouseUpLeft = true
				s.file.LeftTool.MouseUp(int(s.cursor.X), int(s.cursor.Y))
			}
		}

		if rl.IsMouseButtonDown(rl.MouseRightButton) {
			if s.file.HasDoneMouseUpRight {
				s.file.AppendHistory(HistoryAction{make(map[IntVec2]PixelStateData), s.file.CurrentLayer})
			}
			s.file.HasDoneMouseUpRight = false
			s.file.RightTool.MouseDown(int(s.cursor.X), int(s.cursor.Y))
		} else {
			if s.file.HasDoneMouseUpRight == false {
				s.file.HasDoneMouseUpRight = true
				s.file.RightTool.MouseUp(int(s.cursor.X), int(s.cursor.Y))
			}
		}
	}
}
