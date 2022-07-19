package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// PickerTool Pickers an area of the same colored pixels
type PickerTool struct {
	name string
}

// NewPickerTool returns the Picker tool. Requires a name.
func NewPickerTool(name string) *PickerTool {
	return &PickerTool{
		name: name,
	}
}

// MouseDown is for mouse down events
func (t *PickerTool) MouseDown(x, y int32, button MouseButton) {
}

// MouseUp is for mouse up events
func (t *PickerTool) MouseUp(x, y int32, button MouseButton) {
	color, ok := CurrentFile.GetCurrentLayer().PixelData[IntVec2{x, y}]
	if ok {
		switch button {
		case rl.MouseLeftButton:
			CurrentColorSetLeftColor(color)
			SetUIColors(color)
		case rl.MouseRightButton:
			CurrentColorSetRightColor(color)
			SetUIColors(color)
		}
	}

}

// DrawPreview is for drawing the preview
func (t *PickerTool) DrawPreview(x, y int32) {
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
func (t *PickerTool) DrawUI(camera rl.Camera2D) {

}

func (t *PickerTool) String() string {
	return t.name
}
