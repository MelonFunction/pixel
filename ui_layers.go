package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	currentLayerHoverable *Hoverable
	interactables         = make(map[int]*Entity)
)

func LayersUISetCurrentLayer(index int) {
	currentLayerHoverable.Selected = false
	entity, ok := interactables[index]
	if ok {
		if res, err := scene.QueryID(entity.ID); err == nil {
			hoverable := res.Components[entity.Scene.ComponentsMap["hoverable"]].(*Hoverable)
			interactable := res.Components[entity.Scene.ComponentsMap["interactable"]].(*Interactable)

			interactable.OnMouseUp(entity, rl.MouseLeftButton)

			currentLayerHoverable = hoverable
		}
	}
}

func NewLayersUI(bounds rl.Rectangle, file *File) *Entity {
	var buttonHeight float32 = 48.0
	var list *Entity

	makeBox := func(y int, name string) *Entity {
		hidden := NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight), "./res/icons/eye_open.png", false,
			func(entity *Entity, button rl.MouseButton) {
				// button up
				if res, err := scene.QueryID(entity.ID); err == nil {
					drawable := res.Components[entity.Scene.ComponentsMap["drawable"]].(*Drawable)
					// hoverable := res.Components[entity.Scene.ComponentsMap["hoverable"]].(*Hoverable)
					file.Layers[y].Hidden = !file.Layers[y].Hidden

					drawableTexture, ok := drawable.DrawableType.(*DrawableTexture)
					if ok {
						if file.Layers[y].Hidden {
							drawableTexture.SetTexture("./res/icons/eye_closed.png")
						} else {
							drawableTexture.SetTexture("./res/icons/eye_open.png")
						}
					}
				}
			}, nil)
		isCurrent := file.CurrentLayer == y
		label := NewButtonText(rl.NewRectangle(buttonHeight, 0, bounds.Width-buttonHeight*2, buttonHeight), name, isCurrent,
			func(entity *Entity, button rl.MouseButton) {
				// button up
				if res, err := scene.QueryID(entity.ID); err == nil {
					hoverable := res.Components[entity.Scene.ComponentsMap["hoverable"]].(*Hoverable)

					if currentLayerHoverable != nil {
						currentLayerHoverable.Selected = false
					}
					currentLayerHoverable = hoverable
					hoverable.Selected = true

					file.SetCurrentLayer(y)
				}
			}, nil)

		// Set current layer ref
		if res, err := scene.QueryID(label.ID); err == nil {
			hoverable := res.Components[label.Scene.ComponentsMap["hoverable"]].(*Hoverable)

			if isCurrent {
				currentLayerHoverable = hoverable
			}

			interactables[y] = label
		}

		box := NewBox(rl.NewRectangle(0, 0, bounds.Width, buttonHeight), []*Entity{
			hidden,
			label,
		}, FlowDirectionHorizontal)
		return box
	}

	// New layer button
	newLayerButton := NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight), "./res/icons/plus.png", false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			file.AddNewLayer()
			max := len(file.Layers)
			last := file.Layers[max-2] // ignore the temp layer

			if currentLayerHoverable != nil {
				currentLayerHoverable.Selected = false
			}

			list.PushChild(makeBox(max-2, last.Name))
			list.FlowChildren()
		}, nil)

	list = NewScrollableList(rl.NewRectangle(0, buttonHeight, bounds.Width, bounds.Height-buttonHeight), []*Entity{}, FlowDirectionVerticalReversed)
	// All of the layers
	for i, layer := range file.Layers {
		if i == len(file.Layers)-1 {
			// ignore hidden layer
			continue
		}
		list.PushChild(makeBox(i, layer.Name))
	}

	container := NewBox(bounds, []*Entity{
		newLayerButton,
		list,
	}, FlowDirectionVertical)
	return container
}
