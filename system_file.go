package main

import (
	"fmt"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	// ShowDebug enables debug overlays when true
	ShowDebug = false
)

type UIFileSystem struct {
	BasicSystem

	UI map[string]*Entity

	Camera rl.Camera2D
	target rl.Vector2

	// Used for relational mouse movement
	mouseX, mouseY, mouseLastX, mouseLastY int

	cursor rl.Vector2
}

func NewUIFileSystem() *UIFileSystem {
	f := &UIFileSystem{
		Camera: rl.Camera2D{Zoom: 8.0},

		UI: map[string]*Entity{
			"menu": NewMenuUI(rl.NewRectangle(
				0,
				0,
				float32(rl.GetScreenWidth()),
				UIFontSize*20,
			)),
			"editors": NewEditorsUI(rl.NewRectangle(
				0,
				UIFontSize*2,
				float32(rl.GetScreenWidth()),
				UIFontSize*2)),
			"rgb": NewRGBUI(rl.NewRectangle(
				float32(rl.GetScreenWidth()-128*1.5),
				float32(0),
				128*1.5,
				128*1.8)),
			"palette": NewPaletteUI(rl.NewRectangle(
				float32(rl.GetScreenWidth()-int(UIButtonHeight)*2-128*1.5),
				float32(0),
				UIButtonHeight*2,
				128*1.8-UIButtonHeight)),
			"currentColor": NewCurrentColorUI(rl.NewRectangle(
				float32(rl.GetScreenWidth()-int(UIButtonHeight)*2-128*1.5),
				float32(128*1.8-UIButtonHeight*0.66),
				UIButtonHeight*2,
				UIButtonHeight)),
			"tools": NewToolsUI(rl.NewRectangle(
				float32(rl.GetScreenWidth()-int(UIButtonHeight)*2-128*1.5),
				128*1.8,
				128*1.5,
				UIButtonHeight)),
			"layers": NewLayersUI(rl.NewRectangle(
				float32(rl.GetScreenWidth()-128*2.5),
				float32(rl.GetScreenHeight()-128*2),
				128*2.5,
				128*2)),
			"resize": NewResizeUI(),
		},
	}

	f.Camera.Offset.X = float32(rl.GetScreenWidth()) / 2
	f.Camera.Offset.Y = float32(rl.GetScreenHeight()) / 2

	return f
}

func (s *UIFileSystem) Draw() {
	layer := CurrentFile.GetCurrentLayer()

	// Draw
	rl.BeginTextureMode(layer.Canvas)
	if !layer.hasInitialFill {
		CurrentFile.ClearBackground(layer.InitialFillColor)
		layer.hasInitialFill = true
	}
	rl.EndTextureMode()

	// Draw temp layer
	rl.BeginTextureMode(CurrentFile.Layers[len(CurrentFile.Layers)-1].Canvas)
	// LeftTool draws last as it's more important
	if rl.IsMouseButtonDown(rl.MouseRightButton) {
		CurrentFile.RightTool.DrawPreview(int(s.cursor.X), int(s.cursor.Y))

	} else {
		CurrentFile.LeftTool.DrawPreview(int(s.cursor.X), int(s.cursor.Y))
	}

	rl.EndTextureMode()

	// Draw layers
	rl.BeginMode2D(s.Camera)
	for _, layer := range CurrentFile.Layers {
		if !layer.Hidden {
			rl.DrawTextureRec(layer.Canvas.Texture,
				rl.NewRectangle(0, 0, float32(layer.Canvas.Texture.Width), -float32(layer.Canvas.Texture.Height)),
				rl.NewVector2(-float32(layer.Canvas.Texture.Width)/2, -float32(layer.Canvas.Texture.Height)/2),
				rl.White)
		}
	}

	// Grid drawing
	// TODO use a high resolution texture to draw grids, then we won't need to draw lines each draw call
	for x := 0; x <= CurrentFile.CanvasWidth; x += CurrentFile.TileWidth {
		rl.DrawLine(
			-CurrentFile.CanvasWidth/2+x,
			-CurrentFile.CanvasHeight/2,
			-CurrentFile.CanvasWidth/2+x,
			CurrentFile.CanvasHeight/2,
			rl.White)
	}
	for y := 0; y <= CurrentFile.CanvasHeight; y += CurrentFile.TileHeight {
		rl.DrawLine(
			-CurrentFile.CanvasWidth/2,
			-CurrentFile.CanvasHeight/2+y,
			CurrentFile.CanvasWidth/2,
			-CurrentFile.CanvasHeight/2+y,
			rl.White)
	}
	rl.EndMode2D()

	// Debug text
	if ShowDebug {
		for y, history := range CurrentFile.History {
			str := fmt.Sprintf("Layer: %d, Diff: %d",
				history.LayerIndex,
				len(history.PixelState))
			rl.DrawText(str, 20, 20*y+260, 20, rl.White)
		}

		rl.DrawText(fmt.Sprintf("UIInteractableCapturedInput: %v", UIInteractableCapturedInput), 0, 160, 20, rl.White)
		rl.DrawText(fmt.Sprintf("UIEntityCapturedInput: %v", UIEntityCapturedInput), 0, 180, 20, rl.White)
		rl.DrawText(fmt.Sprintf("Current layer: %d", CurrentFile.CurrentLayer), 0, 200, 20, rl.White)
		rl.DrawText(fmt.Sprintf("HistoryOffset: %d", CurrentFile.historyOffset), 0, 220, 20, rl.White)
		rl.DrawText(fmt.Sprintf("History Len: %d", len(CurrentFile.History)), 0, 240, 20, rl.White)
	}
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
				case "tools":
					fallthrough
				case "palette":
					fallthrough
				case "currentColor":
					moveable.Bounds.X = float32(rl.GetScreenWidth() - int(UIButtonHeight)*2 - 128*1.5)
					entity.FlowChildren()

				}
			}

		}

	}

	layer := CurrentFile.GetCurrentLayer()
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
			if CurrentFile.HasDoneMouseUpLeft {
				// Create new history action
				CurrentFile.AppendHistory(HistoryAction{PixelChangeHistoryType, make(map[IntVec2]PixelStateData), CurrentFile.CurrentLayer})
			}
			CurrentFile.HasDoneMouseUpLeft = false

			// Repeated action
			CurrentFile.LeftTool.MouseDown(int(s.cursor.X), int(s.cursor.Y), rl.MouseLeftButton)
		} else {
			// Always fires once
			if CurrentFile.HasDoneMouseUpLeft == false {
				CurrentFile.HasDoneMouseUpLeft = true
				CurrentFile.LeftTool.MouseUp(int(s.cursor.X), int(s.cursor.Y), rl.MouseLeftButton)
			}
		}

		if rl.IsMouseButtonDown(rl.MouseRightButton) {
			if CurrentFile.HasDoneMouseUpRight {
				CurrentFile.AppendHistory(HistoryAction{PixelChangeHistoryType, make(map[IntVec2]PixelStateData), CurrentFile.CurrentLayer})
			}
			CurrentFile.HasDoneMouseUpRight = false
			CurrentFile.RightTool.MouseDown(int(s.cursor.X), int(s.cursor.Y), rl.MouseRightButton)
		} else {
			if CurrentFile.HasDoneMouseUpRight == false {
				CurrentFile.HasDoneMouseUpRight = true
				CurrentFile.RightTool.MouseUp(int(s.cursor.X), int(s.cursor.Y), rl.MouseRightButton)
			}
		}
	}
}
