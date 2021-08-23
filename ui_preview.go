package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	previewContainer *Entity
)

// PreviewUIDrawTile draws the tile in the preview
// TODO add buttons for preview modes
func PreviewUIDrawTile(x, y int) {

	clampedPos := GetClampedCoordinates(x, y)
	tilePos := GetTilePosition(clampedPos.X, clampedPos.Y)

	drawable, ok := previewContainer.GetDrawable()
	if ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			rl.BeginTextureMode(renderTexture.Texture)
			rl.ClearBackground(rl.Black)

			rl.DrawTexturePro(
				CurrentFile.GetCurrentLayer().Canvas.Texture,
				// rl.NewRectangle(0, 0, float32(CurrentFile.CanvasWidth), -float32(CurrentFile.CanvasHeight)),
				rl.NewRectangle(
					float32(tilePos.X),
					-float32(tilePos.Y)-float32(CurrentFile.TileHeight),
					float32(CurrentFile.TileWidth),
					-float32(CurrentFile.TileHeight)),
				rl.NewRectangle(0, 0, float32(renderTexture.Texture.Texture.Width), float32(renderTexture.Texture.Texture.Height)),
				rl.NewVector2(0, 0),
				0,
				rl.White,
			)

			rl.EndTextureMode()
		}
	}
}

// NewPreviewUI creates the UI for previewing the current animation/tile
func NewPreviewUI(bounds rl.Rectangle) *Entity {
	previewContainer = NewRenderTexture(bounds, nil, nil)
	drawable, ok := previewContainer.GetDrawable()
	if ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			rl.BeginTextureMode(renderTexture.Texture)
			rl.ClearBackground(rl.Red)
			rl.EndTextureMode()
		}
	}

	return previewContainer
}
