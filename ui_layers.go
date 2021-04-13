package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	currentLayerHoverable *Hoverable
	interactables         = make(map[int]*Entity)

	list          *Entity
	listContainer *Entity
)

// LayersUISetCurrentLayer can be used to activate a callback on a layer button
// Intended to be used by the ControlSystem
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

func LayersUIMakeList(bounds rl.Rectangle) {
	list = NewScrollableList(rl.NewRectangle(0, UIButtonHeight, bounds.Width, bounds.Height-UIButtonHeight), []*Entity{}, FlowDirectionVerticalReversed)
	// All of the layers
	for i, layer := range CurrentFile.Layers {
		if i == len(CurrentFile.Layers)-1 {
			// ignore hidden layer
			continue
		}
		list.PushChild(LayersUIMakeBox(i, layer))
	}
	list.FlowChildren()
}

func LayersUIRebuildList() {
	list.DestroyNested()
	list.Destroy()
	listContainer.RemoveChild(list)

	if res, err := scene.QueryID(listContainer.ID); err == nil {
		moveable := res.Components[listContainer.Scene.ComponentsMap["moveable"]].(*Moveable)
		bounds := moveable.Bounds
		LayersUIMakeList(bounds)
		listContainer.PushChild(list)
		listContainer.FlowChildren()
	}
}

func LayersUIMakeBox(y int, layer *Layer) *Entity {
	var bounds rl.Rectangle
	if res, err := scene.QueryID(listContainer.ID); err == nil {
		moveable := res.Components[listContainer.Scene.ComponentsMap["moveable"]].(*Moveable)
		bounds = moveable.Bounds
	}

	hidden := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight), "./res/icons/eye_open.png", false,
		func(entity *Entity, button rl.MouseButton) {
			// button up

			if res, err := scene.QueryID(entity.ID); err == nil {
				drawable := res.Components[entity.Scene.ComponentsMap["drawable"]].(*Drawable)
				// hoverable := res.Components[entity.Scene.ComponentsMap["hoverable"]].(*Hoverable)
				CurrentFile.Layers[y].Hidden = !CurrentFile.Layers[y].Hidden

				drawableTexture, ok := drawable.DrawableType.(*DrawableTexture)
				if ok {
					if CurrentFile.Layers[y].Hidden {
						drawableTexture.SetTexture("./res/icons/eye_closed.png")
					} else {
						drawableTexture.SetTexture("./res/icons/eye_open.png")
					}
				}
			}
		}, nil)

	preview := NewRenderTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight), nil, nil)
	if res, err := scene.QueryID(preview.ID); err == nil {
		drawable := res.Components[preview.Scene.ComponentsMap["drawable"]].(*Drawable)
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			renderTexture.Texture = layer.Canvas
		}
	}

	isCurrent := CurrentFile.CurrentLayer == y
	label := NewInput(rl.NewRectangle(0, 0, bounds.Width-UIButtonHeight*3, UIButtonHeight), layer.Name, isCurrent,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			if res, err := scene.QueryID(entity.ID); err == nil {
				hoverable := res.Components[entity.Scene.ComponentsMap["hoverable"]].(*Hoverable)

				if currentLayerHoverable != nil {
					currentLayerHoverable.Selected = false
				}
				currentLayerHoverable = hoverable
				hoverable.Selected = true

				CurrentFile.SetCurrentLayer(y)
			}
		}, nil,
		func(entity *Entity, key rl.Key) {
			// key pressed
			if res, err := scene.QueryID(entity.ID); err == nil {
				drawable := res.Components[entity.Scene.ComponentsMap["drawable"]].(*Drawable)
				drawableParent, ok := drawable.DrawableType.(*DrawableText)
				if ok {
					switch {
					case key >= 97 && key <= 97+26:
						fallthrough
					case key >= rl.KeyA && key <= rl.KeyZ:
						drawableParent.Label += string(rune(key))
					case key == rl.KeyBackspace:
						drawableParent.Label = drawableParent.Label[:len(drawableParent.Label)-1]
					}
				}
			}

		})

	// Set current layer ref
	if res, err := scene.QueryID(label.ID); err == nil {
		hoverable := res.Components[label.Scene.ComponentsMap["hoverable"]].(*Hoverable)

		if isCurrent {
			currentLayerHoverable = hoverable
		}

		interactables[y] = label
	}

	box := NewBox(rl.NewRectangle(0, 0, bounds.Width, UIButtonHeight), []*Entity{
		hidden,
		preview,
		label,
	}, FlowDirectionHorizontal)
	return box
}

// NewLayersUI creates the UI representation of the CurrentFile's layers
func NewLayersUI(bounds rl.Rectangle) *Entity {
	// New layer button
	newLayerButton := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight), "./res/icons/plus.png", false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			CurrentFile.AddNewLayer()
			max := len(CurrentFile.Layers)
			last := CurrentFile.Layers[max-2] // ignore the temp layer

			if currentLayerHoverable != nil {
				currentLayerHoverable.Selected = false
			}

			list.PushChild(LayersUIMakeBox(max-2, last))
			list.FlowChildren()
		}, nil)

	listContainer = NewBox(bounds, []*Entity{
		newLayerButton,
	}, FlowDirectionVertical)

	LayersUIMakeList(bounds)
	listContainer.PushChild(list)
	listContainer.FlowChildren()

	return listContainer
}
