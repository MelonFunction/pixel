package main

import (
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// vars
var (
	PaletteUIPaletteEntity *Entity
	paletteName            *Entity

	// The palette item being dragged
	movingColor *Moveable

	PaletteUICurrentColorEntity *Entity
	PaletteUINextColorEntity    *Entity
	PaletteUIPrevColorEntity    *Entity

	// Triangle denoting the currently selected color
	// This will hide when the color is changed in the color picker or the
	// color is deleted
	currentColorIndicatorEntity *Entity
)

// PaletteUIRemoveColor removes an color from the palette
func PaletteUIRemoveColor(child *Entity) {
	PaletteUIPaletteEntity.RemoveChild(child)
	PaletteUIPaletteEntity.FlowChildren()
}

// PaletteUIUpdateCurrentColorIndicator moves the currentColorIndicatorEntity
// to the position of the currently selected color
func PaletteUIUpdateCurrentColorIndicator() {
	// Create
	if currentColorIndicatorEntity == nil {
		if PaletteUICurrentColorEntity == nil {
			return
		}
		cm, ok := PaletteUICurrentColorEntity.GetMoveable()
		if !ok {
			return
		}
		currentColorIndicatorEntity = NewRenderTexture(cm.Bounds, nil, nil)

		if r, ok := currentColorIndicatorEntity.GetResizeable(); ok {
			r.OnResize = func(entity *Entity) {
				PaletteUIUpdateCurrentColorIndicator()
			}
		}
	}

	// Move and recolor
	if PaletteUICurrentColorEntity != nil {
		cm, ok := PaletteUICurrentColorEntity.GetMoveable()
		if !ok {
			return
		}
		im, ok := currentColorIndicatorEntity.GetMoveable()
		if !ok {
			return
		}

		currentColorIndicatorEntity.Show()

		if t, ok := currentColorIndicatorEntity.GetDrawable(); ok {
			cc := LeftColor
			if tex, ok := t.DrawableType.(*DrawableRenderTexture); ok {
				rl.BeginTextureMode(tex.Texture)
				rl.ClearBackground(rl.Blank)
				rl.DrawTriangle(
					rl.NewVector2(0, 0),
					rl.NewVector2(0, 0+cm.Bounds.Height/2),
					rl.NewVector2(0+cm.Bounds.Width/2, 0),
					rl.NewColor(cc.R+128, cc.G+128, cc.B+128, 255),
				)
				rl.EndTextureMode()
			}
		}

		// move indicator on scroll
		im.Bounds.X = cm.Bounds.X
		im.Bounds.Y = cm.Bounds.Y + cm.Offset.Y
	}

	// Show on top
	currentColorIndicatorEntity.Scene.MoveEntityToEnd(currentColorIndicatorEntity)
}

// PaletteUIHideCurrentColorIndicator hides the currentColorIndicatorEntity
func PaletteUIHideCurrentColorIndicator() {
	if currentColorIndicatorEntity != nil {
		currentColorIndicatorEntity.Hide()

		if PaletteUICurrentColorEntity != nil {
			PaletteUINextColorEntity = PaletteUICurrentColorEntity
			PaletteUIPrevColorEntity = PaletteUICurrentColorEntity
		}
	}
}

// PaletteUINextColor selects the next color
func PaletteUINextColor() {
	if PaletteUINextColorEntity != nil {
		PaletteUICurrentColorEntity = PaletteUINextColorEntity
		if i, ok := PaletteUICurrentColorEntity.GetInteractable(); ok {
			i.OnMouseUp(PaletteUICurrentColorEntity, rl.MouseLeftButton)
		}
		PaletteUIUpdateCurrentColorIndicator()
	}
}

// PaletteUIPreviousColor selects the previous color
func PaletteUIPreviousColor() {
	if PaletteUIPrevColorEntity != nil {
		PaletteUICurrentColorEntity = PaletteUIPrevColorEntity
		if i, ok := PaletteUICurrentColorEntity.GetInteractable(); ok {
			i.OnMouseUp(PaletteUICurrentColorEntity, rl.MouseLeftButton)
		}
		PaletteUIUpdateCurrentColorIndicator()
	}
}

// PaletteUIRebuildPalette rebuilds the current palette
func PaletteUIRebuildPalette() {
	PaletteUIPrevColorEntity = nil
	PaletteUINextColorEntity = nil
	PaletteUICurrentColorEntity = nil
	PaletteUIHideCurrentColorIndicator()

	if drawable, ok := paletteName.GetDrawable(); ok {
		if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
			drawableText.Label = Settings.PaletteData[CurrentFile.CurrentPalette].Name
		}

		if children, err := PaletteUIPaletteEntity.GetChildren(); err == nil {
			for i := len(children) - 1; i >= 0; i-- {
				PaletteUIPaletteEntity.RemoveChild(children[i])
			}
		}

		PaletteUIPrevColorEntity = nil
		for i, color := range Settings.PaletteData[CurrentFile.CurrentPalette].data {
			c := PaletteUIAddColor(color, int32(i))
			if i == 0 {
				PaletteUICurrentColorEntity = c
			} else if i == 1 {
				PaletteUINextColorEntity = c
			}
		}
	}
	PaletteUIUpdateCurrentColorIndicator()
}

// PaletteUIAddColor adds a color to the palette
func PaletteUIAddColor(color rl.Color, index int32) *Entity {
	var w float32
	var h float32
	if res, err := scene.QueryID(PaletteUIPaletteEntity.ID); err == nil {
		moveable := res.Components[PaletteUIPaletteEntity.Scene.ComponentsMap["moveable"]].(*Moveable)
		w = moveable.Bounds.Width / 5
		h = moveable.Bounds.Width / 5
	}

	var e *Entity
	e = NewRenderTexture(rl.NewRectangle(0, 0, w, h),
		func(entity *Entity, button MouseButton) {
			// Up
			switch button {
			case rl.MouseLeftButton:
				CurrentColorSetLeftColor(color)
				// SetUIColors(color)
				PaletteUICurrentColorEntity = entity

				children, err := PaletteUIPaletteEntity.GetChildren()
				if err != nil {
					log.Println(err)
					return
				}

				// Get the element the cursor is over
				moveToPosition := 0
				// The index of the dragged child
				childPosition := 0
				// isMoveBefore is true if the cursor was on the left
				// half of the item
				isMoveBefore := true

				collision := false
				for i, child := range children {
					if child == entity {
						childPosition = i

						PaletteUINextColorEntity = nil
						PaletteUIPrevColorEntity = nil
						if i+1 < len(children) {
							PaletteUINextColorEntity = children[i+1]
						}
						if i-1 >= 0 {
							PaletteUIPrevColorEntity = children[i-1]
						}
						PaletteUIUpdateCurrentColorIndicator()
					} else {
						if res, err := scene.QueryID(child.ID); err == nil {
							childMoveable := res.Components[child.Scene.ComponentsMap["moveable"]].(*Moveable)
							cur := rl.GetMousePosition()
							bounds := childMoveable.Bounds
							if rl.CheckCollisionPointRec(cur, bounds) {
								collision = true
								moveToPosition = i
								isMoveBefore = cur.X < (bounds.X + bounds.Width/2)
							}
						}
					}
				}

				if movingColor == nil {
					return
				}

				movingColor = nil

				if collision {
					moved := children[childPosition]
					movedData := Settings.PaletteData[CurrentFile.CurrentPalette].data[childPosition]
					children = append(children[:childPosition], children[childPosition+1:]...)
					Settings.PaletteData[CurrentFile.CurrentPalette].data = append(Settings.PaletteData[CurrentFile.CurrentPalette].data[:childPosition], Settings.PaletteData[CurrentFile.CurrentPalette].data[childPosition+1:]...)
					if childPosition < moveToPosition {
						moveToPosition--
					}
					if isMoveBefore == false {
						moveToPosition++
					}
					children = append(children[:moveToPosition], append([]*Entity{moved}, children[moveToPosition:]...)...)
					Settings.PaletteData[CurrentFile.CurrentPalette].data = append(
						Settings.PaletteData[CurrentFile.CurrentPalette].data[:moveToPosition],
						append(
							[]rl.Color{movedData}, Settings.PaletteData[CurrentFile.CurrentPalette].data[moveToPosition:]...)...)
					SaveSettings()
				}
				PaletteUIPaletteEntity.FlowChildren()
			case rl.MouseRightButton:
				// SetUIColors(color)
				CurrentColorSetRightColor(color)
			}
		},
		func(entity *Entity, button MouseButton, isHeld bool) {
			// Down
			if isHeld {
				switch button {
				case rl.MouseLeftButton:
					if movingColor == nil {
						if moveable, ok := entity.GetMoveable(); ok {
							movingColor = moveable
						}
					}

					movingColor.Bounds.X = rl.GetMousePosition().X - movingColor.Bounds.Width/2
					movingColor.Bounds.Y = rl.GetMousePosition().Y - movingColor.Bounds.Height/2
				}
			}
		})
	if res, err := scene.QueryID(e.ID); err == nil {
		drawable := res.Components[e.Scene.ComponentsMap["drawable"]].(*Drawable)
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			texture := renderTexture.Texture
			rl.BeginTextureMode(texture)
			rl.ClearBackground(color)
			rl.EndTextureMode()
		}
	}
	if moveable, ok := e.GetMoveable(); ok {
		moveable.Draggable = true
	}

	PaletteUIPaletteEntity.PushChild(e)
	PaletteUIPaletteEntity.FlowChildren()

	return e
}

// NewPaletteUI returns a new PaletteUI
func NewPaletteUI(bounds rl.Rectangle) *Entity {
	PaletteUIPaletteEntity = NewScrollableList(rl.NewRectangle(0, 0, bounds.Width, bounds.Height-UIButtonHeight/2), []*Entity{}, FlowDirectionHorizontal)

	paletteName = NewInput(rl.NewRectangle(0, 0, bounds.Width, UIButtonHeight/2),
		Settings.PaletteData[CurrentFile.CurrentPalette].Name,
		TextAlignCenter,
		false, func(entity *Entity, button MouseButton) {}, nil,
		func(entity *Entity, key Key) {
			if drawable, ok := entity.GetDrawable(); ok {
				if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
					if key == rl.KeyEnter {
						RemoveCapturedInput()
					} else if key == rl.KeyBackspace && len(drawableText.Label) > 0 {
						drawableText.Label = drawableText.Label[:len(drawableText.Label)-1]
					} else if len(drawableText.Label) < 8 {
						drawableText.Label += string(rune(key))
					}

					Settings.PaletteData[CurrentFile.CurrentPalette].Name = drawableText.Label
				}
			}
		})
	if interactable, ok := paletteName.GetInteractable(); ok {
		interactable.OnBlur = func(entity *Entity) {
			SaveSettings()
		}
	}

	paletteContainer := NewBox(bounds, []*Entity{
		paletteName,
		PaletteUIPaletteEntity,
	}, FlowDirectionVertical)

	PaletteUIRebuildPalette()
	PaletteUIUpdateCurrentColorIndicator()

	if interactable, ok := PaletteUIPaletteEntity.GetInteractable(); ok {
		interactable.OnScroll = func(direction int32) {
			PaletteUIUpdateCurrentColorIndicator()
		}
	}

	return paletteContainer
}
