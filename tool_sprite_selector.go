package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// SpriteSelectorTool selects tiles
type SpriteSelectorTool struct {
	name      string
	firstDown bool    // if mouse has been pressed
	firstPos  IntVec2 // the first tile selected
	lastPos   IntVec2 // the last tile selected

	firstSprite, lastSprite int32 // sprite sheet position of the selected sprites
	onMouseUp               func(firstSprite, lastSprite int32)
}

// NewSpriteSelectorTool returns the fill tool. Requires a name.
func NewSpriteSelectorTool(name string, onMouseUp func(firstSprite, lastSprite int32)) *SpriteSelectorTool {
	return &SpriteSelectorTool{
		name:      name,
		onMouseUp: onMouseUp,
	}
}

// MouseDown is for mouse down events
func (t *SpriteSelectorTool) MouseDown(x, y int32, button MouseButton) {
	clampedPos := GetClampedCoordinates(x, y)
	tilePos := GetTilePosition(clampedPos.X, clampedPos.Y)
	sheetPos := tilePos.X/CurrentFile.TileWidth + (tilePos.Y/CurrentFile.TileHeight)*(CurrentFile.CanvasWidth/CurrentFile.TileWidth)

	if t.firstDown == false {
		t.firstDown = true
		t.firstPos = tilePos

		t.firstSprite = sheetPos
	}

	t.lastPos = tilePos
	t.lastSprite = sheetPos

	// Don't let the lastPos be in front of the firstPos
	if t.firstPos.Y > t.lastPos.Y || (t.firstPos.Y >= t.lastPos.Y && t.firstPos.X > t.lastPos.X) {
		t.firstDown = false
	}
}

// MouseUp is for mouse up events
func (t *SpriteSelectorTool) MouseUp(x, y int32, button MouseButton) {
	switch button {
	case rl.MouseLeftButton:
	case rl.MouseRightButton:
	}

	t.onMouseUp(t.firstSprite, t.lastSprite)

	t.firstDown = false

}

// DrawPreview is for drawing the preview
func (t *SpriteSelectorTool) DrawPreview(x, y int32) {
	rl.ClearBackground(rl.Blank)

	if t.firstDown {
		rl.DrawRectangle(t.firstPos.X, t.firstPos.Y, CurrentFile.TileWidth/2, CurrentFile.TileHeight, rl.Orange)
		rl.DrawRectangle(t.lastPos.X+CurrentFile.TileWidth/2, t.lastPos.Y, CurrentFile.TileWidth/2, CurrentFile.TileHeight, rl.Blue)
	} else {
		// Preview pixel location with a suitable color
		color := rl.NewColor(255, 255, 255, 192)
		pos := GetTilePosition(x, y)
		rl.DrawRectangle(pos.X, pos.Y, CurrentFile.TileWidth, CurrentFile.TileHeight, color)
	}
}

// DrawUI is for drawing the UI
func (t *SpriteSelectorTool) DrawUI(camera rl.Camera2D) {

}

func (t *SpriteSelectorTool) String() string {
	return t.name
}
