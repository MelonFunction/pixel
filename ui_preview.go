package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	previewContainer        *Entity
	previewButtonsContainer *Entity
	previewArea             *Entity
	currentPreviewMode      previewMode
	previewZoom             int // how much preview is zoomed

	previewCurrentButton          *Entity
	previewCurrentSheetButton     *Entity
	previewCurrentTileButton      *Entity
	previewCurrentAnimationButton *Entity
	previewCurrentPixelButton     *Entity
)

type previewMode int

// Preview modes
const (
	previewCurrentSheet     previewMode = iota // shows the entire spritesheet, can zoom
	previewCurrentTile                         // shows the current sprite, tiled
	previewCurrentAnimation                    // shows the current animation
	previewCurrentPixel                        // follows mouse cursor around
)

// PreviewUIDrawTile draws the tile in the preview
// TODO add buttons for preview modes
func PreviewUIDrawTile(x, y int) {

	drawable, ok := previewArea.GetDrawable()
	if ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			rl.BeginTextureMode(renderTexture.Texture)
			rl.ClearBackground(rl.Black)

			switch currentPreviewMode {
			case previewCurrentSheet:
				rl.DrawTexturePro(
					CurrentFile.GetCurrentLayer().Canvas.Texture,
					// rl.NewRectangle(0, 0, float32(CurrentFile.CanvasWidth), -float32(CurrentFile.CanvasHeight)),
					rl.NewRectangle(
						0,
						0,
						float32(CurrentFile.CanvasWidth),
						-float32(CurrentFile.CanvasHeight)),
					rl.NewRectangle(0, 0, float32(renderTexture.Texture.Texture.Width), float32(renderTexture.Texture.Texture.Height)),
					rl.NewVector2(0, 0),
					0,
					rl.White,
				)

			case previewCurrentTile:
				clampedPos := GetClampedCoordinates(x, y)
				tilePos := GetTilePosition(clampedPos.X, clampedPos.Y)

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
			case previewCurrentAnimation:
			case previewCurrentPixel:
			}

			rl.EndTextureMode()
		}
	}
}

// NewPreviewUI creates the UI for previewing the current animation/tile
func NewPreviewUI(bounds rl.Rectangle) *Entity {
	previewArea = NewRenderTexture(bounds, nil, nil)
	drawable, ok := previewArea.GetDrawable()
	if ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			rl.BeginTextureMode(renderTexture.Texture)
			rl.ClearBackground(rl.Red)
			rl.EndTextureMode()
		}
	}

	unselectCurrentButton := func() {
		if previewCurrentButton == nil {
			return
		}
		hoverable, ok := previewCurrentButton.GetHoverable()
		if ok {
			hoverable.Selected = false
		}
	}

	selectCurrentButton := func() {
		if previewCurrentButton == nil {
			return
		}
		hoverable, ok := previewCurrentButton.GetHoverable()
		if ok {
			hoverable.Selected = true
		}
	}

	// buttons
	previewCurrentSheetButton = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/currentSheet.png"), false, func(entity *Entity, button rl.MouseButton) {
			currentPreviewMode = previewCurrentSheet
			unselectCurrentButton()
			previewCurrentButton = previewCurrentSheetButton
			selectCurrentButton()
		}, nil)
	previewCurrentTileButton = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/currentTile.png"), false, func(entity *Entity, button rl.MouseButton) {
			currentPreviewMode = previewCurrentTile
			unselectCurrentButton()
			previewCurrentButton = previewCurrentTileButton
			selectCurrentButton()
		}, nil)
	previewCurrentAnimationButton = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/currentAnimation.png"), false, func(entity *Entity, button rl.MouseButton) {
			currentPreviewMode = previewCurrentAnimation
			unselectCurrentButton()
			previewCurrentButton = previewCurrentAnimationButton
			selectCurrentButton()
		}, nil)
	previewCurrentPixelButton = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/currentPixel.png"), false, func(entity *Entity, button rl.MouseButton) {
			currentPreviewMode = previewCurrentPixel
			unselectCurrentButton()
			previewCurrentButton = previewCurrentPixelButton
			selectCurrentButton()
		}, nil)

	previewCurrentButton = previewCurrentSheetButton
	selectCurrentButton()

	previewButtonsContainer = NewBox(
		rl.NewRectangle(
			bounds.X,
			bounds.Y,
			bounds.Width,
			UIButtonHeight,
		),
		[]*Entity{
			previewCurrentSheetButton,
			previewCurrentTileButton,
			previewCurrentAnimationButton,
			previewCurrentPixelButton,
		},
		FlowDirectionHorizontal,
	)

	bounds.Height += UIButtonHeight
	previewContainer = NewBox(bounds, []*Entity{
		previewArea,
		previewButtonsContainer,
	}, FlowDirectionVertical)

	previewContainer.FlowChildren()

	return previewContainer
}
