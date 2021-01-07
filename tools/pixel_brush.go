package tools

import rl "github.com/lachee/raylib-goplus/raylib"

type PixelBrushTool struct {
	lastPos                IntVec2
	shouldConnectToLastPos bool
	Color                  rl.Color
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
	t.Color = color
}
func (t *PixelBrushTool) GetColor() rl.Color {
	return t.Color
}
func (t *PixelBrushTool) DrawPreview(x, y int) {
	rl.ClearBackground(rl.Transparent)
	rl.DrawPixel(x, y, t.GetColor())
}
