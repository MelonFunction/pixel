package main

import (
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
			case rl.MouseRightButton:
				CurrentFile.RightColor = color
				CurrentColorSetColor(currentColorRight, CurrentFile.RightColor)
			case rl.MouseMiddleButton:
				PaletteUIRemoveColor(e)
			}
		},
		func(entity *Entity, button rl.MouseButton, isHeld bool) {
			// Down
			switch button {
			case rl.MouseLeftButton:
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
