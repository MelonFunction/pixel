package main

import rl "github.com/lachee/raylib-goplus/raylib"

func NewRGBUI(bounds rl.Rectangle) *Entity {
	rgb := NewTexture(bounds,
		func(entity *Entity, button rl.MouseButton) {
			// button up
		},
		func(entity *Entity, button rl.MouseButton) {
			// button down
		})
	container := NewBox(bounds, []*Entity{rgb}, FlowDirectionVertical)
	return container
}
