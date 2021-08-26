package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

// PixelBrushTool draws a single pixel at a time and can also double as an
// eraser if eraser is true
type PixelBrushTool struct {
	lastPos                IntVec2
	name                   string
	eraser                 bool
	shouldConnectToLastPos bool
	size                   int // brush size
	// Don't draw over the same pixel multiple times, prevents opacity stacking
	drawnPixels []IntVec2

	currentColor rl.Color
}

// NewPixelBrushTool returns the pixel brush tool. Requires a name and whether
// the tool is in eraser mode (helpful to prevent the current color from being
// lost)
func NewPixelBrushTool(name string, eraser bool) *PixelBrushTool {
	return &PixelBrushTool{
		name:        name,
		eraser:      eraser,
		size:        3,
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

// GetSize returns the tool size
func (t *PixelBrushTool) GetSize() int {
	return t.size
}

// SetSize sets the tool size
func (t *PixelBrushTool) SetSize(size int) {
	if size > 0 {
		t.size = size
	}
}

func (t *PixelBrushTool) drawPixel(x, y int, fileDraw bool) {
	var min, max int
	if t.size%2 == 0 {
		min = -t.size / 2
		max = t.size/2 - 1
	} else {
		min = -t.size / 2
		max = t.size / 2
	}
	for xx := min; xx <= max; xx++ {
		for yy := min; yy <= max; yy++ {
			if fileDraw {
				CurrentFile.DrawPixel(x+xx, y+yy, t.currentColor, true)
				t.drawnPixels = append(t.drawnPixels, IntVec2{x + xx, y + yy})
			} else {
				rl.DrawPixel(x+xx, y+yy, rl.Color{255, 255, 255, 192})
			}
		}
	}
}

func (t *PixelBrushTool) isLineModifierDown() bool {
	for _, keys := range Settings.KeymapData["drawLine"] {
		allDown := true
		for _, key := range keys {
			if !rl.IsKeyDown(key) {
				allDown = false
			}
		}

		if allDown {
			return true
		}
	}
	return false
}

// MouseDown is for mouse down events
func (t *PixelBrushTool) MouseDown(x, y int, button rl.MouseButton) {
	// Assume we are in eraser mode by setting transparent as default
	t.currentColor = rl.Transparent
	if !t.eraser {
		switch button {
		case rl.MouseLeftButton:
			t.currentColor = CurrentFile.LeftColor
		case rl.MouseRightButton:
			t.currentColor = CurrentFile.RightColor
		}
	}

	if t.shouldConnectToLastPos || t.isLineModifierDown() {
		Line(t.lastPos.X, t.lastPos.Y, x, y, func(x, y int) {
			if !t.exists(IntVec2{x, y}) {
				t.drawPixel(x, y, true)
			}
		})
	} else {
		t.shouldConnectToLastPos = true
		if !t.exists(IntVec2{x, y}) {
			t.drawPixel(x, y, true)
		}
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

	if t.isLineModifierDown() {
		Line(t.lastPos.X, t.lastPos.Y, x, y, func(x, y int) {
			loc := IntVec2{x, y}
			if !t.exists(loc) {
				t.drawPixel(x, y, false)
			}
		})
	}

	t.drawPixel(x, y, false)
}

func (t *PixelBrushTool) String() string {
	return t.name
}
