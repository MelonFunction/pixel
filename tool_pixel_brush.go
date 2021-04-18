package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

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

func (t *PixelBrushTool) MouseUp(x, y int, button rl.MouseButton) {
	t.shouldConnectToLastPos = false
}

func (t *PixelBrushTool) DrawPreview(x, y int) {
	rl.ClearBackground(rl.Transparent)
	// Don't call file.DrawPixel as history isn't needed for this action
	rl.DrawPixel(x, y, rl.Color{255, 255, 255, 128})
}

func (t *PixelBrushTool) SetFileReference(file *File) {

}

func (t *PixelBrushTool) String() string {
	return t.name
}
