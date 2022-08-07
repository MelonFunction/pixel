package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// FillTool fills an area of the same colored pixels
type FillTool struct {
	lastPos IntVec2
	name    string
	eraser  bool
}

// NewFillTool returns the fill tool. Requires a name.
func NewFillTool(name string) *FillTool {
	return &FillTool{
		name: name,
	}
}

// MouseDown is for mouse down events
func (t *FillTool) MouseDown(x, y int32, button MouseButton) {
}

// MouseUp is for mouse up events
func (t *FillTool) MouseUp(x, y int32, button MouseButton) {
	color := rl.Blank
	switch button {
	case rl.MouseLeftButton:
		color = LeftColor
	case rl.MouseRightButton:
		color = RightColor
	}

	pd := CurrentFile.GetCurrentLayer().PixelData
	clickedColor := pd[IntVec2{x, y}]

	var recFill func(rx, ry int32)
	recFill = func(rx, ry int32) {
		if pd[IntVec2{rx, ry}] == clickedColor && color != clickedColor {
			// Set color
			oldColor := pd[IntVec2{rx, ry}]
			// pd[IntVec2{rx, ry}] = color
			CurrentFile.DrawPixel(rx, ry, color, CurrentFile.GetCurrentLayer())

			// Append history
			if oldColor != color {
				latestHistoryInterface := CurrentFile.History[len(CurrentFile.History)-1]
				latestHistory, ok := latestHistoryInterface.(HistoryPixel)
				if ok {
					ps := latestHistory.PixelState[IntVec2{rx, ry}]
					ps.Current = color
					ps.Prev = oldColor
					latestHistory.PixelState[IntVec2{rx, ry}] = ps
				}
			}

			if rx+1 < CurrentFile.CanvasWidth {
				recFill(rx+1, ry)
			}
			if rx-1 >= 0 {
				recFill(rx-1, ry)
			}
			if ry+1 < CurrentFile.CanvasHeight {
				recFill(rx, ry+1)
			}
			if ry-1 >= 0 {
				recFill(rx, ry-1)
			}
		}
	}
	recFill(x, y)
	// CurrentFile.GetCurrentLayer().Redraw()
	// CurrentFile.RedrawRenderLayer()
}

// DrawPreview is for drawing the preview
func (t *FillTool) DrawPreview(x, y int32) {
	rl.ClearBackground(rl.Blank)
	// Preview pixel location with a suitable color
	c := CurrentFile.GetCurrentLayer().PixelData[IntVec2{x, y}]
	avg := (c.R + c.G + c.B) / 3
	if avg > 255/2 {
		rl.DrawPixel(x, y, rl.NewColor(0, 0, 0, 192))
	} else {
		rl.DrawPixel(x, y, rl.NewColor(255, 255, 255, 192))
	}
}

// DrawUI is for drawing the UI
func (t *FillTool) DrawUI(camera rl.Camera2D) {

}

func (t *FillTool) String() string {
	return t.name
}
