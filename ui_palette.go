package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	paletteEntity *Entity
	paletteName   *Entity

	// The palette item being dragged
	movingColor *Moveable

	// Currently selected color
	currentColorEntity *Entity
	// Triangle denoting the currently selected color
	// This will hide when the color is changed in the color picker or the
	// color is deleted
	currentColorIndicatorEntity *Entity
)

// PaletteUIRemoveColor removes an color from the palette
func PaletteUIRemoveColor(child *Entity) {
	paletteEntity.RemoveChild(child)
	paletteEntity.FlowChildren()
}

// PaletteUIUpdateCurrentColorIndicator moves the currentColorIndicatorEntity
// to the position of the currently selected color
func PaletteUIUpdateCurrentColorIndicator() {
	// Create
	if currentColorIndicatorEntity == nil {
		if currentColorEntity == nil {
			return
		}
		cm, ok := currentColorEntity.GetMoveable()
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
	if currentColorEntity != nil {
		cm, ok := currentColorEntity.GetMoveable()
		if !ok {
			return
		}
		im, ok := currentColorIndicatorEntity.GetMoveable()
		if !ok {
			return
		}

		currentColorIndicatorEntity.Show()

		if t, ok := currentColorIndicatorEntity.GetDrawable(); ok {
			cc := CurrentFile.LeftColor
			if tex, ok := t.DrawableType.(*DrawableRenderTexture); ok {
				rl.BeginTextureMode(tex.Texture)
				rl.ClearBackground(rl.Transparent)
				rl.DrawTriangle(
					rl.NewVector2(0, 0),
					rl.NewVector2(0, 0+cm.Bounds.Height/2),
					rl.NewVector2(0+cm.Bounds.Width/2, 0),
					rl.NewColor(cc.R+128, cc.G+128, cc.B+128, 255),
				)
				rl.EndTextureMode()
			}
		}

		im.Bounds = cm.Bounds
	}

	// Show on top
	currentColorIndicatorEntity.Scene.MoveEntityToEnd(currentColorIndicatorEntity)
}

// PaletteUIHideCurrentColorIndicator hides the currentColorIndicatorEntity
func PaletteUIHideCurrentColorIndicator() {
	if currentColorIndicatorEntity != nil {
		currentColorIndicatorEntity.Hide()
	}
}

// PaletteUIRebuildPalette rebuilds the current palette
func PaletteUIRebuildPalette() {
	if drawable, ok := paletteName.GetDrawable(); ok {
		if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
			drawableText.Label = Settings.PaletteData[CurrentFile.CurrentPalette].Name
		}

		if children, err := paletteEntity.GetChildren(); err == nil {
			for i := len(children) - 1; i >= 0; i-- {
				paletteEntity.RemoveChild(children[i])
			}
		}

		for i, color := range Settings.PaletteData[CurrentFile.CurrentPalette].data {
			c := PaletteUIAddColor(color, i)
			if i == 0 {
				currentColorEntity = c
			}
		}
	}
	PaletteUIUpdateCurrentColorIndicator()
}

// PaletteUIAddColor adds a color to the palette
func PaletteUIAddColor(color rl.Color, index int) *Entity {
	var w float32
	var h float32
	if res, err := scene.QueryID(paletteEntity.ID); err == nil {
		moveable := res.Components[paletteEntity.Scene.ComponentsMap["moveable"]].(*Moveable)
		w = moveable.Bounds.Width / 5
		h = moveable.Bounds.Width / 5
	}

	var e *Entity
	e = NewRenderTexture(rl.NewRectangle(0, 0, w, h),
		func(entity *Entity, button rl.MouseButton) {
			movingColor = nil
			// Up
			switch button {
			case rl.MouseLeftButton:
				CurrentColorSetLeftColor(color)
				SetUIColors(color)
				currentColorEntity = entity

				children, err := paletteEntity.GetChildren()
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
				paletteEntity.FlowChildren()
				PaletteUIUpdateCurrentColorIndicator()
			case rl.MouseRightButton:
				SetUIColors(color)
				CurrentColorSetRightColor(color)
			case rl.MouseMiddleButton:
				// TODO Hold shift to change the "add color to palette (+) button" to "remove color from palette (-) button"

				// PaletteUIRemoveColor(e)
				// Settings.PaletteData[CurrentFile.CurrentPalette].data = append(
				// 	Settings.PaletteData[CurrentFile.CurrentPalette].data[:index],
				// 	Settings.PaletteData[CurrentFile.CurrentPalette].data[index+1:]...,
				// )
				// SaveSettings()
			}
		},
		func(entity *Entity, button rl.MouseButton, isHeld bool) {
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

	paletteEntity.PushChild(e)
	paletteEntity.FlowChildren()

	return e
}

// NewPaletteUI returns a new PaletteUI
func NewPaletteUI(bounds rl.Rectangle) *Entity {
	paletteEntity = NewScrollableList(rl.NewRectangle(0, 0, bounds.Width, bounds.Height-UIButtonHeight/2), []*Entity{}, FlowDirectionHorizontal)

	for i, color := range Settings.PaletteData[CurrentFile.CurrentPalette].data {
		c := PaletteUIAddColor(color, i)
		if i == 0 {
			currentColorEntity = c
		}
	}

	paletteName = NewInput(rl.NewRectangle(0, 0, bounds.Width, UIButtonHeight/2),
		Settings.PaletteData[CurrentFile.CurrentPalette].Name,
		TextAlignCenter,
		false, func(entity *Entity, button rl.MouseButton) {}, nil,
		func(entity *Entity, key rl.Key) {
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
		paletteEntity,
	}, FlowDirectionVertical)

	PaletteUIUpdateCurrentColorIndicator()

	return paletteContainer
}
