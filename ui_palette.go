package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	paletteEntity        *Entity
	selectedPaletteColor *Entity
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
		w = moveable.Bounds.Width / 3
		h = moveable.Bounds.Width / 3
	}

	var e *Entity
	e = NewRenderTexture(rl.NewRectangle(0, 0, w, h),
		func(entity *Entity, button rl.MouseButton) {
			// Up
			switch button {
			case rl.MouseLeftButton:
				CurrentFile.LeftColor = color
				CurrentColorSetColor(currentColorLeft, CurrentFile.LeftColor)

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
					children = append(children[:childPosition], children[childPosition+1:]...)
					if childPosition < moveToPosition {
						moveToPosition--
					}
					if isMoveBefore == false {
						moveToPosition++
					}
					children = append(children[:moveToPosition], append([]*Entity{moved}, children[moveToPosition:]...)...)
				}
				paletteEntity.FlowChildren()
			case rl.MouseRightButton:
				CurrentFile.RightColor = color
				CurrentColorSetColor(currentColorRight, CurrentFile.RightColor)
			case rl.MouseMiddleButton:
				PaletteUIRemoveColor(e)
			}
		},
		func(entity *Entity, button rl.MouseButton, isHeld bool) {
			// Down
			if isHeld {
				switch button {
				case rl.MouseLeftButton:
					if res, err := scene.QueryID(entity.ID); err == nil {
						moveable := res.Components[entity.Scene.ComponentsMap["moveable"]].(*Moveable)
						moveable.Bounds.X = rl.GetMousePosition().X - moveable.Bounds.Width/2
						moveable.Bounds.Y = rl.GetMousePosition().Y - moveable.Bounds.Height/2
					}
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

	paletteEntity.PushChild(e)
	paletteEntity.FlowChildren()
}

func NewPaletteUI(bounds rl.Rectangle) *Entity {
	paletteEntity = NewScrollableList(bounds, []*Entity{}, FlowDirectionHorizontal)
	PaletteUIAddColor(rl.Red)
	PaletteUIAddColor(rl.Blue)
	PaletteUIAddColor(rl.Green)
	PaletteUIAddColor(rl.Pink)
	PaletteUIAddColor(rl.Orange)
	PaletteUIAddColor(rl.Purple)
	PaletteUIAddColor(rl.Aqua)

	return paletteEntity
}
