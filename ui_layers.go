package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

type LayersUI struct {
	file   *File
	Bounds rl.Rectangle

	ButtonWidth, ButtonHeight int
}

func NewLayersUI(bounds rl.Rectangle, file *File) *Entity {
	l := &LayersUI{
		file:         file,
		Bounds:       bounds,
		ButtonWidth:  int(bounds.Width) - 20 - 16,
		ButtonHeight: 20,
	}
	var buttonHeight float32 = 32.0
	var list *Entity

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
		}, FlowDirectionHorizontal)
		return box
	}

	// New layer button
	newLayerButton := NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight), "./res/icons/plus.png", false,
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

	list = NewScrollableList(rl.NewRectangle(0, buttonHeight, l.Bounds.Width, l.Bounds.Height-buttonHeight), []*Entity{}, FlowDirectionVerticalReversed)
	// All of the layers
	for i, layer := range l.file.Layers {
		if i == len(l.file.Layers)-1 {
			continue
		}

		list.PushChild(makeBox(i, layer.Name))
	}

	container := NewBox(l.Bounds, []*Entity{
		newLayerButton,
		list,
	}, FlowDirectionVertical)
	return container
}
