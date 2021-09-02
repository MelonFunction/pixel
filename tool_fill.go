package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
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
func (t *FillTool) MouseDown(x, y int, button rl.MouseButton) {
}

// MouseUp is for mouse up events
func (t *FillTool) MouseUp(x, y int, button rl.MouseButton) {
	color := rl.Transparent
	switch button {
	case rl.MouseLeftButton:
		color = CurrentFile.LeftColor
	case rl.MouseRightButton:
		color = CurrentFile.RightColor
	}

	pd := CurrentFile.GetCurrentLayer().PixelData
	clickedColor := pd[IntVec2{x, y}]

	var recFill func(rx, ry int)
	recFill = func(rx, ry int) {
		if pd[IntVec2{rx, ry}] == clickedColor && color != clickedColor {
			CurrentFile.DrawPixel(rx, ry, color, true)
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
}

// DrawPreview is for drawing the preview
func (t *FillTool) DrawPreview(x, y int) {
	rl.ClearBackground(rl.Transparent)
	// Preview pixel location with a suitable color
	c := CurrentFile.GetCurrentLayer().PixelData[IntVec2{x, y}]
	avg := (c.R + c.G + c.B) / 3
	if avg > 255/2 {
		rl.DrawPixel(x, y, rl.NewColor(0, 0, 0, 192))
	} else {
		rl.DrawPixel(x, y, rl.NewColor(255, 255, 255, 192))
	}
}

func (t *FillTool) String() string {
	return t.name
}
