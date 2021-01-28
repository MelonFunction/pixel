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

	var currentLayerHoverable *Hoverable

	makeBox := func(y int, name string) *Entity {
		hidden := NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight), "./res/icons/eye_open.png", false,
			func(entity *Entity, button rl.MouseButton) {
				// button up
				if res, err := scene.QueryID(entity.ID); err == nil {
					drawable := res.Components[entity.Scene.ComponentsMap["drawable"]].(*Drawable)
					// hoverable := res.Components[entity.Scene.ComponentsMap["hoverable"]].(*Hoverable)
					l.file.Layers[y].Hidden = !l.file.Layers[y].Hidden

					drawableTexture, ok := drawable.DrawableType.(*DrawableTexture)
					if ok {
						if l.file.Layers[y].Hidden {
							drawableTexture.SetTexture("./res/icons/eye_closed.png")
						} else {
							drawableTexture.SetTexture("./res/icons/eye_open.png")
						}
					}
				}
			},
			func(entity *Entity, button rl.MouseButton) {
				// button down
			})
		isCurrent := l.file.CurrentLayer == y
		label := NewButtonText(rl.NewRectangle(buttonHeight, 0, l.Bounds.Width-buttonHeight*2, buttonHeight), name, isCurrent,
			func(entity *Entity, button rl.MouseButton) {
				// button up
				if res, err := scene.QueryID(entity.ID); err == nil {
					hoverable := res.Components[entity.Scene.ComponentsMap["hoverable"]].(*Hoverable)

					log.Println("hoverable", currentLayerHoverable)
					if currentLayerHoverable != nil {
						currentLayerHoverable.Selected = false
					}

					l.file.SetCurrentLayer(y)
					hoverable.Selected = true
					currentLayerHoverable = hoverable
				}
			},
			func(entity *Entity, button rl.MouseButton) {
				// button down
			})
		if isCurrent {
			// Set current layer ref
			if res, err := scene.QueryID(label.ID); err == nil {
				hoverable := res.Components[label.Scene.ComponentsMap["hoverable"]].(*Hoverable)
				currentLayerHoverable = hoverable
			}
		}

		box := NewBox(rl.NewRectangle(0, float32(y)*buttonHeight, l.Bounds.Width, buttonHeight), []*Entity{
			hidden,
			label,
		})
		return box
	}

	// New layer button
	NewButtonTexture(rl.NewRectangle(float32(l.Position.X), float32(l.Position.Y)-buttonHeight, buttonHeight, buttonHeight), "./res/icons/plus.png", false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			l.file.AddNewLayer()
			max := len(l.file.Layers)
			last := l.file.Layers[max-2] // ignore the temp layer

			if currentLayerHoverable != nil {
				currentLayerHoverable.Selected = false
			}

			if res, err := scene.QueryID(entity.ID); err == nil {
				hoverable := res.Components[entity.Scene.ComponentsMap["hoverable"]].(*Hoverable)
				hoverable.Selected = true
				currentLayerHoverable = hoverable
			}

			list.PushChild(makeBox(max-2, last.Name))
			list.FlowChildren()
		},
		func(entity *Entity, button rl.MouseButton) {
			// button down
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

	list = NewScrollableList(rl.NewRectangle(float32(l.Position.X), float32(l.Position.Y), l.Bounds.Width, l.Bounds.Height), layers, true)
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
		UIElementWithControl = nil
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
}

func (l *LayersUI) Destroy() {
}
