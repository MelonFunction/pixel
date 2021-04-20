package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	currentColorEntity *Entity // the container

	currentColorLeft  *Entity
	currentColorRight *Entity
	currentColorSwap  *Entity
	currentColorAdd   *Entity
)

func CurrentColorSetLeftColor(color rl.Color) {
	if drawable, ok := currentColorLeft.GetDrawable(); ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			CurrentFile.LeftColor = color

			texture := renderTexture.Texture
			rl.BeginTextureMode(texture)
			rl.ClearBackground(color)
			rl.EndTextureMode()
		}
	}
}

func CurrentColorSetRightColor(color rl.Color) {
	if drawable, ok := currentColorRight.GetDrawable(); ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			CurrentFile.RightColor = color

			texture := renderTexture.Texture
			rl.BeginTextureMode(texture)
			rl.ClearBackground(color)
			rl.EndTextureMode()
		}
	}
}

func CurrentColorUIAddColor(color rl.Color) *Entity {
	var w float32
	var h float32
	if res, err := scene.QueryID(currentColorEntity.ID); err == nil {
		moveable := res.Components[currentColorEntity.Scene.ComponentsMap["moveable"]].(*Moveable)
		w = moveable.Bounds.Width / 4
		h = moveable.Bounds.Width / 4
	}

	e := NewRenderTexture(rl.NewRectangle(0, 0, w, h), nil, nil)

	currentColorEntity.PushChild(e)
	currentColorEntity.FlowChildren()
	return e
}

func NewCurrentColorUI(bounds rl.Rectangle) *Entity {
	currentColorEntity = NewBox(bounds, []*Entity{}, FlowDirectionHorizontal)

	currentColorLeft = CurrentColorUIAddColor(CurrentFile.LeftColor)
	CurrentColorSetLeftColor(CurrentFile.LeftColor)
	currentColorRight = CurrentColorUIAddColor(CurrentFile.RightColor)
	CurrentColorSetRightColor(CurrentFile.RightColor)

	currentColorSwap = NewButtonTexture(rl.NewRectangle(0, 0, bounds.Width/4, bounds.Width/4), "./res/icons/swap.png", false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			left := CurrentFile.LeftColor
			right := CurrentFile.RightColor
			CurrentColorSetLeftColor(right)
			CurrentColorSetRightColor(left)
		}, nil)
	currentColorEntity.PushChild(currentColorSwap)

	currentColorAdd = NewButtonTexture(rl.NewRectangle(0, 0, bounds.Width/4, bounds.Width/4), "./res/icons/plus.png", false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			PaletteUIAddColor(CurrentFile.LeftColor)
		}, nil)
	currentColorEntity.PushChild(currentColorAdd)
	currentColorEntity.FlowChildren()

	return currentColorEntity
}
