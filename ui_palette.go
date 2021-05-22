package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	paletteEntity        *Entity
	selectedPaletteColor *Entity

	// The palette item being dragged
	movingColor *Moveable
)

func PaletteUIRemoveColor(child *Entity) {
	paletteEntity.RemoveChild(child)
	paletteEntity.FlowChildren()
}

func PaletteUIAddColor(color rl.Color) {
	var w float32
	var h float32
	if res, err := scene.QueryID(paletteEntity.ID); err == nil {
		moveable := res.Components[paletteEntity.Scene.ComponentsMap["moveable"]].(*Moveable)
		w = moveable.Bounds.Width / 4
		h = moveable.Bounds.Width / 4
	}

	// Get the element the cursor is over
	moveToPosition := 0
	// The index of the dragged child
	childPosition := 0
	// isMoveBefore is true if the cursor was on the left
	// half of the item
	isMoveBefore := true
	// if there was a collision
	collision := false

	var e *Entity
	e = NewRenderTexture(rl.NewRectangle(0, 0, w, h),
		func(entity *Entity, button rl.MouseButton) {
			movingColor = nil
			// Up
			switch button {
			case rl.MouseLeftButton:
				CurrentColorSetLeftColor(color)
				SetUIColors(color)

				SaveSettings()
				paletteEntity.FlowChildren()
			case rl.MouseRightButton:
				SetUIColors(color)
				CurrentColorSetRightColor(color)
			case rl.MouseMiddleButton:
				PaletteUIRemoveColor(e)
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

					switch button {
					case rl.MouseLeftButton:

						children, err := paletteEntity.GetChildren()
						if err != nil {
							log.Println(err)
							return
						}

						collision = false
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
										break
									}
								}
							}

						}

						if collision {
							moved := children[childPosition]
							movedData := Settings.PaletteData[0].Data[childPosition]
							children = append(children[:childPosition], children[childPosition+1:]...)
							Settings.PaletteData[0].Data = append(Settings.PaletteData[0].Data[:childPosition], Settings.PaletteData[0].Data[childPosition+1:]...)
							if childPosition < moveToPosition {
								moveToPosition--
							}
							if isMoveBefore == false {
								moveToPosition++
							}
							children = append(children[:moveToPosition], append([]*Entity{moved}, children[moveToPosition:]...)...)

							// TODO get current palette
							Settings.PaletteData[0].Data = append(
								Settings.PaletteData[0].Data[:moveToPosition],
								append(
									[]rl.Color{movedData}, Settings.PaletteData[0].Data[moveToPosition:]...)...)
						}
						paletteEntity.FlowChildren()
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
}

func NewPaletteUI(bounds rl.Rectangle) *Entity {
	paletteEntity = NewScrollableList(bounds, []*Entity{}, FlowDirectionHorizontal)
	for _, color := range Settings.PaletteData[0].Data {
		PaletteUIAddColor(color)
	}

	return paletteEntity
}
