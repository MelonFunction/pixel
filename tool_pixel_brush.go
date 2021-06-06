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
	// Don't draw over the same pixel multiple times, prevents opacity stacking
	drawnPixels []IntVec2
}

// NewPixelBrushTool returns the pixel brush tool. Requires a name and whether
// the tool is in eraser mode (helpful to prevent the current color from being
// lost)
func NewPixelBrushTool(name string, eraser bool) *PixelBrushTool {
	return &PixelBrushTool{
		name:        name,
		eraser:      eraser,
		drawnPixels: make([]IntVec2, 0, 100),
	}
}

func (t *PixelBrushTool) exists(e IntVec2) bool {
	for _, v := range t.drawnPixels {
		if v == e {
			return true
		}
	}

	return false
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

	for _, keys := range Settings.KeymapData["drawLine"] {
		allDown := true
		for _, key := range keys {
			if !rl.IsKeyDown(key) {
				allDown = false
			}
		}

		if allDown {
			t.shouldConnectToLastPos = true
			break
		}
	}

	if !t.shouldConnectToLastPos {
		t.shouldConnectToLastPos = true
		loc := IntVec2{x, y}
		if !t.exists(loc) {
			t.drawnPixels = append(t.drawnPixels, loc)
			CurrentFile.DrawPixel(x, y, color, true)
		}
	} else {
		Line(t.lastPos.X, t.lastPos.Y, x, y, func(x, y int) {
			loc := IntVec2{x, y}
			if !t.exists(loc) {
				CurrentFile.DrawPixel(x, y, color, true)
				t.drawnPixels = append(t.drawnPixels, loc)
			}
		})
	}
	t.lastPos.X = x
	t.lastPos.Y = y
}

// MouseUp is for mouse up events
func (t *PixelBrushTool) MouseUp(x, y int, button rl.MouseButton) {
	t.shouldConnectToLastPos = false
	t.drawnPixels = make([]IntVec2, 0, 100)
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
