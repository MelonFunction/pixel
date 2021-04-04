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

func CurrentColorSetColor(entity *Entity, color rl.Color) {
	if res, err := scene.QueryID(entity.ID); err == nil {
		drawable := res.Components[entity.Scene.ComponentsMap["drawable"]].(*Drawable)
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
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
		w = moveable.Bounds.Width / 3
		h = moveable.Bounds.Width / 3
	}

	e := NewRenderTexture(rl.NewRectangle(0, 0, w, h), nil, nil)
	CurrentColorSetColor(e, color)

	currentColorEntity.PushChild(e)
	currentColorEntity.FlowChildren()
	return e
}

func NewCurrentColorUI(bounds rl.Rectangle) *Entity {
	currentColorEntity = NewBox(bounds, []*Entity{}, FlowDirectionHorizontal)

	currentColorLeft = CurrentColorUIAddColor(CurrentFile.LeftColor)
	currentColorRight = CurrentColorUIAddColor(CurrentFile.RightColor)

	// currentColorSwap = NewButtonTexture(rl.NewRectangle(0, 0, bounds.Width/3, bounds.Width/3), "./res/icons/plus.png", false,
	// 	func(entity *Entity, button rl.MouseButton, isHeld bool) {
	// 		// button up
	// 		CurrentFile.LeftColor, CurrentFile.RightColor = CurrentFile.RightColor, CurrentFile.LeftColor
	// 		CurrentColorSetColor(currentColorLeft, CurrentFile.LeftColor)
	// 		CurrentColorSetColor(currentColorRight, CurrentFile.RightColor)
	// 	}, nil)
	// currentColorEntity.PushChild(currentColorSwap)
	// currentColorEntity.FlowChildren()

	currentColorAdd = NewButtonTexture(rl.NewRectangle(0, 0, bounds.Width/3, bounds.Width/3), "./res/icons/plus.png", false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			PaletteUIAddColor(CurrentFile.LeftColor)
		}, nil)
	currentColorEntity.PushChild(currentColorAdd)
	currentColorEntity.FlowChildren()

	return currentColorEntity
}
