package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

type LayersUI struct {
	Position      IntVec2
	file          *File
	name          string
	Width, Height int
	Bounds        rl.Rectangle

	ButtonWidth, ButtonHeight int

	Texture rl.RenderTexture2D

	wasMouseButtonDown bool
}

func NewLayersUI(position IntVec2, width, height int, file *File, name string) *LayersUI {
	l := &LayersUI{
		Position:     position,
		file:         file,
		name:         name,
		Width:        width,
		Height:       height,
		Bounds:       rl.NewRectangle(float32(position.X), float32(position.Y), float32(width), float32(height)),
		ButtonWidth:  width - 20 - 16,
		ButtonHeight: 20,
		// Texture:            rl.LoadRenderTexture(width, height),
		wasMouseButtonDown: false,
	}
	l.generateUI()
	return l
}

func (l *LayersUI) generateUI() {
	var buttonHeight float32 = 32.0
	var list *Entity
	layers := make([]*Entity, 0, 16)

	makeBox := func(y int, name string) *Entity {
		return NewBox(rl.NewRectangle(0, float32(y)*buttonHeight, l.Bounds.Width, buttonHeight), []*Entity{
			NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight), "./res/icons/plus.png", false,
				func(button rl.MouseButton) {
					l.file.Layers[y].Hidden = !l.file.Layers[y].Hidden
				},
				func(button rl.MouseButton) {

				}),
			NewButtonText(rl.NewRectangle(buttonHeight, 0, l.Bounds.Width-buttonHeight*2, buttonHeight), name, false,
				func(button rl.MouseButton) {
					l.file.SetCurrentLayer(y)
				},
				func(button rl.MouseButton) {

				}),
		})
	}

	// New layer button
	NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight), "./res/icons/plus.png", false,
		func(button rl.MouseButton) {
			l.file.AddNewLayer()
			max := len(l.file.Layers)
			last := l.file.Layers[max-2]

			list.PushChild(makeBox(max-2, last.Name))
			list.FlowChildren()
		},
		func(button rl.MouseButton) {

		})

	// All of the layers
	for i, layer := range l.file.Layers {
		if i == len(l.file.Layers)-1 {
			continue
		}

		var box *Entity
		box = makeBox(i, layer.Name)
		layers = append(layers, box)
	}

	list = NewScrollableList(rl.NewRectangle(32, 32, l.Bounds.Width, l.Bounds.Height), layers, true)
	list.FlowChildren()

}

func (l *LayersUI) GetWasMouseButtonDown() bool {
	return l.wasMouseButtonDown
}

func (l *LayersUI) SetWasMouseButtonDown(isDown bool) {
	l.wasMouseButtonDown = isDown
}

func (l *LayersUI) MouseUp() {
	if l == UIElementWithControl {
		// UIHasControl = false
		UIElementWithControl = nil
		// UIComponentWithControl = nil // unset child too
	}
}

func (l *LayersUI) MouseDown() {
	// Using update instead since we have to check for hover on each component
}

func (l *LayersUI) CheckCollisions(offset rl.Vector2) bool {
	return false
}

func (l *LayersUI) Update() {
}

func (l *LayersUI) Draw() {
	// rl.BeginTextureMode(l.Texture)
	// rl.ClearBackground(rl.Transparent)

	// for _, component := range l.Components {
	// 	component.Draw()
	// }

	// rl.EndTextureMode()

	// l.scrollbar.Draw()

	// rl.DrawTextureRec(l.Texture.Texture,
	// 	rl.NewRectangle(0, 0, float32(l.Texture.Texture.Width), -float32(l.Texture.Texture.Height)),
	// 	rl.NewVector2(float32(l.Position.X), float32(l.Position.Y)),
	// 	rl.White)
}

func (l *LayersUI) Destroy() {
	// l.box.Destroy()
	// l.Texture.Unload()
	// for _, component := range l.Components {
	// 	component.Destroy()
	// }
	// l.scrollbar.Destroy()
}
