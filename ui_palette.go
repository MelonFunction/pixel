package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	paletteEntity        *Entity
	selectedPaletteColor *Entity
)

func PaletteUIAddColor(color rl.Color) {

	e := NewRenderTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight),
		func(entity *Entity, button rl.MouseButton) {
			switch button {
			case rl.MouseLeftButton:
				CurrentFile.LeftColor = color
			case rl.MouseRightButton:
				CurrentFile.RightColor = color
			}
		}, nil)
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
	return paletteEntity
}
