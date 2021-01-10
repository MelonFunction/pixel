package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

type PixelBrushTool struct {
	lastPos                IntVec2
	shouldConnectToLastPos bool
	color                  rl.Color
	file                   *File
	name                   string
}

func NewPixelBrushTool(color rl.Color, file *File, name string) *PixelBrushTool {
	return &PixelBrushTool{
		color: color,
		file:  file,
		name:  name,
	}
}

func (t *PixelBrushTool) MouseDown(x, y int) {
	if !t.shouldConnectToLastPos {
		t.shouldConnectToLastPos = true
		t.file.DrawPixel(x, y, t.GetColor(), true)
	} else {
		Line(t.lastPos.X, t.lastPos.Y, x, y, func(x, y int) {
			t.file.DrawPixel(x, y, t.GetColor(), true)
		})
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
	// Don't call file.DrawPixel as history isn't needed for this action
	rl.DrawPixel(x, y, t.GetColor())
}
func (t *PixelBrushTool) SetFileReference(file *File) {

}
func (t *PixelBrushTool) String() string {
	return t.name
}
