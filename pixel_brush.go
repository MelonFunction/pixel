package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

type PixelBrushTool struct {
	lastPos                IntVec2
	shouldConnectToLastPos bool
	color                  rl.Color
	file                   *File
}

func NewPixelBrushTool(color rl.Color, file *File) *PixelBrushTool {
	return &PixelBrushTool{
		color: color,
		file:  file,
	}
}

func (t *PixelBrushTool) MouseDown(x, y int) {
	if !t.shouldConnectToLastPos {
		t.shouldConnectToLastPos = true
		rl.DrawPixel(x, y, t.GetColor())
	} else {
		Line(t.lastPos.X, t.lastPos.Y, x, y, t.GetColor())
	}
	t.lastPos.X = x
	t.lastPos.Y = y
}
func (t *PixelBrushTool) MouseUp(x, y int) {
	t.shouldConnectToLastPos = false
}
func (t *PixelBrushTool) SetColor(color rl.Color) {
	t.color = color
}
func (t *PixelBrushTool) GetColor() rl.Color {
	return t.color
}
func (t *PixelBrushTool) DrawPreview(x, y int) {
	rl.ClearBackground(rl.Transparent)
	rl.DrawPixel(x, y, t.GetColor())
}
func (t *PixelBrushTool) SetFileReference(file *File) {

}
