package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

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

func (t *FillTool) MouseDown(x, y int, button rl.MouseButton) {
}

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
		if pd[IntVec2{rx, ry}] == clickedColor {
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

func (t *FillTool) DrawPreview(x, y int) {
	rl.ClearBackground(rl.Transparent)
	// Don't call file.DrawPixel as history isn't needed for this action
	rl.DrawPixel(x, y, rl.Color{255, 255, 255, 128})
}

func (t *FillTool) SetFileReference(file *File) {

}

func (t *FillTool) String() string {
	return t.name
}
