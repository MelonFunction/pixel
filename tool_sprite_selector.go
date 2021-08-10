package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

// SpriteSelectorTool selects tiles
type SpriteSelectorTool struct {
	lastPos IntVec2
	name    string
}

// NewSpriteSelectorTool returns the fill tool. Requires a name.
func NewSpriteSelectorTool(name string) *SpriteSelectorTool {
	return &SpriteSelectorTool{
		name: name,
	}
}

// MouseDown is for mouse down events
func (t *SpriteSelectorTool) MouseDown(x, y int, button rl.MouseButton) {
}

// MouseUp is for mouse up events
func (t *SpriteSelectorTool) MouseUp(x, y int, button rl.MouseButton) {
	switch button {
	case rl.MouseLeftButton:
	case rl.MouseRightButton:
	}

}

// DrawPreview is for drawing the preview
func (t *SpriteSelectorTool) DrawPreview(x, y int) {
	rl.ClearBackground(rl.Transparent)
	// Preview pixel location with a suitable color
	c := CurrentFile.GetCurrentLayer().PixelData[IntVec2{x, y}]
	avg := (c.R + c.G + c.B) / 3
	var color rl.Color
	if avg > 255/2 {
		color = rl.NewColor(0, 0, 0, 192)
	} else {
		color = rl.NewColor(255, 255, 255, 192)
	}
	x = x / CurrentFile.TileWidth * CurrentFile.TileWidth
	y = y / CurrentFile.TileHeight * CurrentFile.TileHeight
	rl.DrawRectangle(x, y, CurrentFile.TileWidth, CurrentFile.TileHeight, color)
}

func (t *SpriteSelectorTool) String() string {
	return t.name
}
