package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

type LayersUI struct {
	Position      IntVec2
	file          *File
	name          string
	Width, Height int
	Bounds        rl.Rectangle

	ButtonWidth, ButtonHeight int
	box                       *Entity
	button                    *Entity

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

	// l.button = NewButtonText(rl.NewRectangle(0, 0, l.Bounds.Width, l.Bounds.Height), "hello", false,
	// 	func(button rl.MouseButton) {
	// 		log.Println("i was clicked", button)
	// 	},
	// 	func(button rl.MouseButton) {

	// 	})
	var buttonHeight float32 = 32.0
	NewScrollableList(rl.NewRectangle(100, 100, l.Bounds.Width, l.Bounds.Height), []*Entity{
		NewBox(rl.NewRectangle(0, 0, l.Bounds.Width, buttonHeight), []*Entity{
			NewButtonText(rl.NewRectangle(buttonHeight, 0, l.Bounds.Width-buttonHeight*2, buttonHeight), "hello", false,
				func(button rl.MouseButton) {
					log.Println("hello button was clicked", button)
				},
				func(button rl.MouseButton) {

				}),
			NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight), "./res/icons/plus.png", false,
				func(button rl.MouseButton) {
					log.Println("world button was clicked", button)
				},
				func(button rl.MouseButton) {

				}),
		}),
	})

	// NewBox(rl.NewRectangle(300, 300, l.Bounds.Width, l.Bounds.Height), []*Entity{
	// 	NewButtonText(rl.NewRectangle(0, 0, l.Bounds.Width, l.Bounds.Height/2), "hello", false,
	// 		func(button rl.MouseButton) {
	// 			log.Println("hello button was clicked", button)
	// 		},
	// 		func(button rl.MouseButton) {

	// 		}),
	// 	NewButtonText(rl.NewRectangle(0, l.Bounds.Height/2, l.Bounds.Width, l.Bounds.Height/2), "world", false,
	// 		func(button rl.MouseButton) {
	// 			log.Println("world button was clicked", button)
	// 		},
	// 		func(button rl.MouseButton) {

	// 		}),
	// })

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
