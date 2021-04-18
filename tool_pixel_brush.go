package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

// PixelBrushTool draws a single pixel at a time and can also double as an
// eraser if eraser is true
type PixelBrushTool struct {
	lastPos                IntVec2
	shouldConnectToLastPos bool
	name                   string
	eraser                 bool
}

// NewPixelBrushTool returns the pixel brush tool. Requires a name and whether
// the tool is in eraser mode (helpful to prevent the current color from being
// lost)
func NewPixelBrushTool(name string, eraser bool) *PixelBrushTool {
	return &PixelBrushTool{
		name:   name,
		eraser: eraser,
	}
}

// MouseDown is for mouse down events
func (t *PixelBrushTool) MouseDown(x, y int, button rl.MouseButton) {
	// Assume we are in eraser mode by setting transparent as default
	color := rl.Transparent
	if !t.eraser {
		switch button {
		case rl.MouseLeftButton:
			color = CurrentFile.LeftColor
		case rl.MouseRightButton:
			color = CurrentFile.RightColor
		}
	}

	if !t.shouldConnectToLastPos {
		t.shouldConnectToLastPos = true
		CurrentFile.DrawPixel(x, y, color, true)
	} else {
		Line(t.lastPos.X, t.lastPos.Y, x, y, func(x, y int) {
			CurrentFile.DrawPixel(x, y, color, true)
		})
	}
	t.lastPos.X = x
	t.lastPos.Y = y
}

// MouseUp is for mouse up events
func (t *PixelBrushTool) MouseUp(x, y int, button rl.MouseButton) {
	t.shouldConnectToLastPos = false
}

// DrawPreview is for drawing the preview
func (t *PixelBrushTool) DrawPreview(x, y int) {
	rl.ClearBackground(rl.Transparent)
	// Preview pixel location with a suitable color
	c := CurrentFile.GetCurrentLayer().PixelData[IntVec2{x, y}]
	avg := (c.R + c.G + c.B) / 3
	if avg > 255/2 {
		rl.DrawPixel(x, y, rl.Color{0, 0, 0, 192})
	} else {
		rl.DrawPixel(x, y, rl.Color{255, 255, 255, 192})
	}
}

func (t *PixelBrushTool) String() string {
	return t.name
}
