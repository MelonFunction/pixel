package main

import (
	"fmt"
	"log"
	"strconv"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	previewContainer                 *Entity
	previewButtonsContainer          *Entity
	previewAnimationButtonsContainer *Entity

	previewArea              *Entity
	currentPreviewMode       previewMode
	previewZoom              int     // how much preview is zoomed
	previewAnimationTimer    float32 // keeps track of time between anim frames
	previewAnimationIsPaused bool    // true if animation is paused
	previewAnimationFrame    int     // current frame of animation, accessed by ui_animations

	previewCurrentButton          *Entity
	previewCurrentSheetButton     *Entity
	previewCurrentTileButton      *Entity
	previewCurrentAnimationButton *Entity
	previewCurrentPixelButton     *Entity
	previewCurrentAnimationTiming *Entity // input which displays the current animation's timing
)

type previewMode int

// Preview modes
const (
	previewCurrentSheet     previewMode = iota // shows the entire spritesheet, can zoom
	previewCurrentTile                         // shows the current sprite, tiled
	previewCurrentPixel                        // follows mouse cursor around
	previewCurrentAnimation                    // shows the current animation
)

// PreviewUISetTiming sets the timing in the preview input
func PreviewUISetTiming(timing float32) {
	if drawable, ok := previewCurrentAnimationTiming.GetDrawable(); ok {
		if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
			drawableText.Label = fmt.Sprintf("%00.f", timing)
			CurrentFile.SetCurrentAnimationTiming(timing)
		}
	}
}

// PreviewUIDrawTile draws the tile in the preview
func PreviewUIDrawTile(x, y int) {
	// TODO squish data from all layers
	drawable, ok := previewArea.GetDrawable()
	if ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			rl.BeginTextureMode(renderTexture.Texture)
			rl.ClearBackground(rl.Black)

			switch currentPreviewMode {
			case previewCurrentSheet:
				for _, layer := range CurrentFile.Layers {
					if !layer.Hidden {
						rl.DrawTexturePro(
							layer.Canvas.Texture,
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
					}
				}

			case previewCurrentTile:
				// TODO button to select and lock the tile being previewed
				clampedPos := GetClampedCoordinates(x, y)
				tilePos := GetTilePosition(clampedPos.X, clampedPos.Y)

				for x := 0; x < 3; x++ {
					for y := 0; y < 3; y++ {
						for i, layer := range CurrentFile.Layers {
							if i == len(CurrentFile.Layers)-1 {
								continue
							}
							if !layer.Hidden {
								rl.DrawTexturePro(
									layer.Canvas.Texture,
									// rl.NewRectangle(0, 0, float32(CurrentFile.CanvasWidth), -float32(CurrentFile.CanvasHeight)),
									rl.NewRectangle(
										float32(tilePos.X),
										-float32(tilePos.Y)-float32(CurrentFile.TileHeight),
										float32(CurrentFile.TileWidth),
										-float32(CurrentFile.TileHeight)),
									rl.NewRectangle(
										float32(renderTexture.Texture.Texture.Width)/3*float32(x),
										float32(renderTexture.Texture.Texture.Height)/3*float32(y),
										float32(renderTexture.Texture.Texture.Width)/3,
										float32(renderTexture.Texture.Texture.Height)/3),
									rl.NewVector2(0, 0),
									0,
									rl.White,
								)
							}
						}
					}
				}

			case previewCurrentPixel:
				clampedPos := GetClampedCoordinates(x, y)

				for i, layer := range CurrentFile.Layers {
					if i == len(CurrentFile.Layers)-1 {
						continue
					}
					if !layer.Hidden {
						rl.DrawTexturePro(
							layer.Canvas.Texture,
							// rl.NewRectangle(0, 0, float32(CurrentFile.CanvasWidth), -float32(CurrentFile.CanvasHeight)),
							rl.NewRectangle(
								float32(clampedPos.X)-float32(CurrentFile.TileWidth)/2,
								-float32(clampedPos.Y)-float32(CurrentFile.TileHeight)/2,
								float32(CurrentFile.TileWidth),
								-float32(CurrentFile.TileHeight)),
							rl.NewRectangle(0, 0, float32(renderTexture.Texture.Texture.Width), float32(renderTexture.Texture.Texture.Height)),
							rl.NewVector2(0, 0),
							0,
							rl.White,
						)
					}
				}

				// Draw 2 rectangles so that the pixel is always highlighted
				// regardless of the color
				cursorSize := int(renderTexture.Texture.Texture.Width) / CurrentFile.TileWidth
				rl.DrawRectangleLinesEx(rl.NewRectangle(
					float32(renderTexture.Texture.Texture.Width)/2+2,
					float32(renderTexture.Texture.Texture.Height)/2+2,
					float32(cursorSize)-4,
					float32(cursorSize)-4,
				),
					2,
					rl.Gray,
				)
				rl.DrawRectangleLinesEx(rl.NewRectangle(
					float32(renderTexture.Texture.Texture.Width)/2,
					float32(renderTexture.Texture.Texture.Height)/2,
					float32(cursorSize),
					float32(cursorSize),
				),
					2,
					rl.White,
				)

			case previewCurrentAnimation:

				anim := CurrentFile.GetCurrentAnimation()
				if !previewAnimationIsPaused {
					previewAnimationTimer += rl.GetFrameTime()
				}
				if anim != nil {
					if previewAnimationTimer > 1.0/anim.Timing {
						// Get next frame
						previewAnimationTimer = 0
						previewAnimationFrame++
						if previewAnimationFrame > anim.FrameEnd {
							previewAnimationFrame = anim.FrameStart
						}
					}
				}
				// Convert tile number to coords
				tilePos := IntVec2{
					X: (previewAnimationFrame * CurrentFile.TileWidth) % CurrentFile.CanvasWidth,
					Y: ((previewAnimationFrame * CurrentFile.TileHeight) / (CurrentFile.CanvasWidth)) * CurrentFile.TileHeight,
				}

				for i, layer := range CurrentFile.Layers {
					if i == len(CurrentFile.Layers)-1 {
						continue
					}
					if !layer.Hidden {
						rl.DrawTexturePro(
							layer.Canvas.Texture,
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
					}
				}

			}

			rl.DrawRectangleLinesEx(rl.NewRectangle(0, 0, float32(renderTexture.Texture.Texture.Width), float32(renderTexture.Texture.Texture.Height)), 2, rl.Gray)

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

		previewAnimationButtonsContainer.Hide()
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

	// Buttons
	previewCurrentSheetButton = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/current_sheet.png"), false, func(entity *Entity, button rl.MouseButton) {
			currentPreviewMode = previewCurrentSheet
			unselectCurrentButton()
			previewCurrentButton = previewCurrentSheetButton
			selectCurrentButton()
		}, nil)
	previewCurrentTileButton = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/current_tile.png"), false, func(entity *Entity, button rl.MouseButton) {
			currentPreviewMode = previewCurrentTile
			unselectCurrentButton()
			previewCurrentButton = previewCurrentTileButton
			selectCurrentButton()
		}, nil)
	previewCurrentPixelButton = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/current_pixel.png"), false, func(entity *Entity, button rl.MouseButton) {
			currentPreviewMode = previewCurrentPixel
			unselectCurrentButton()
			previewCurrentButton = previewCurrentPixelButton
			selectCurrentButton()
		}, nil)
	previewCurrentAnimationButton = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/current_animation.png"), false, func(entity *Entity, button rl.MouseButton) {
			currentPreviewMode = previewCurrentAnimation
			unselectCurrentButton()
			previewCurrentButton = previewCurrentAnimationButton
			selectCurrentButton()
			// Show animation controls
			previewAnimationButtonsContainer.Show()

			// Set starting frame
			anim := CurrentFile.GetCurrentAnimation()
			if anim != nil {
				previewAnimationFrame = anim.FrameStart
			}

		}, nil)

	previewCurrentAnimationTiming = NewInput(rl.NewRectangle(0, 0, UIButtonHeight*1.5, UIButtonHeight/2), "10", TextAlignCenter, false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
		},
		nil,
		func(entity *Entity, key rl.Key) {
			// key pressed
			if drawable, ok := entity.GetDrawable(); ok {
				if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
					// TODO this could probably be added to util since the same
					// code exists in multiple places
					if key == rl.KeyBackspace && len(drawableText.Label) > 0 {
						drawableText.Label = drawableText.Label[:len(drawableText.Label)-1]
					} else if len(drawableText.Label) < 12 {
						switch {
						// 0 to 9
						case key >= 48 && key <= 57:
							fallthrough
						// a to z
						case key >= 97 && key <= 97+26:
							fallthrough
						case key >= rl.KeyA && key <= rl.KeyZ:
							drawableText.Label += string(rune(key))
						}

						fl, err := strconv.ParseFloat(drawableText.Label, 32)
						if err != nil {
							log.Println(err)
						}
						drawableText.Label = fmt.Sprintf("%00.f", fl)
						CurrentFile.SetCurrentAnimationTiming(float32(fl))
					}
				}
			}
		})

	if interactable, ok := previewCurrentAnimationTiming.GetInteractable(); ok {
		interactable.OnScroll = func(direction int) {
			if drawable, ok := previewCurrentAnimationTiming.GetDrawable(); ok {
				if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
					fl, err := strconv.ParseFloat(drawableText.Label, 32)
					if err != nil {
						log.Println(err)
					}
					fl += float64(direction)
					drawableText.Label = fmt.Sprintf("%00.f", fl)
					CurrentFile.SetCurrentAnimationTiming(float32(fl))
				}
			}
		}
	}

	// Animation controls
	previewAnimationButtonsContainer = NewBox(
		rl.NewRectangle(0, 0, UIButtonHeight*1.5, UIButtonHeight),
		[]*Entity{
			NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2),
				GetFile("./res/icons/play_pause.png"), false, func(entity *Entity, button rl.MouseButton) {
					previewAnimationIsPaused = !previewAnimationIsPaused
				}, nil),
			NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2),
				GetFile("./res/icons/arrow_left.png"), false, func(entity *Entity, button rl.MouseButton) {
					anim := CurrentFile.GetCurrentAnimation()
					if anim == nil {
						return
					}
					previewAnimationFrame--
					if previewAnimationFrame < anim.FrameStart {
						previewAnimationFrame = anim.FrameEnd
					}
				}, nil),
			NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2),
				GetFile("./res/icons/arrow_right.png"), false, func(entity *Entity, button rl.MouseButton) {
					anim := CurrentFile.GetCurrentAnimation()
					if anim == nil {
						return
					}
					previewAnimationFrame++
					if previewAnimationFrame > anim.FrameEnd {
						previewAnimationFrame = anim.FrameStart
					}
				}, nil),

			previewCurrentAnimationTiming,
		},
		FlowDirectionHorizontal)
	previewAnimationButtonsContainer.Hide()

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
			previewCurrentPixelButton,
			previewCurrentAnimationButton,
			previewAnimationButtonsContainer,
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
