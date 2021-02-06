package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	buttonHeight float32 = 48.0

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
	list = NewScrollableList(rl.NewRectangle(0, buttonHeight, bounds.Width, bounds.Height-buttonHeight), []*Entity{}, FlowDirectionVerticalReversed)
	// All of the layers
	for i, layer := range CurrentFile.Layers {
		if i == len(CurrentFile.Layers)-1 {
			// ignore hidden layer
			continue
		}
		list.PushChild(LayersUIMakeBox(i, layer.Name))
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

func LayersUIMakeBox(y int, name string) *Entity {
	var bounds rl.Rectangle
	if res, err := scene.QueryID(listContainer.ID); err == nil {
		moveable := res.Components[listContainer.Scene.ComponentsMap["moveable"]].(*Moveable)
		bounds = moveable.Bounds
	}

	hidden := NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight), "./res/icons/eye_open.png", false,
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
	isCurrent := CurrentFile.CurrentLayer == y
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

				CurrentFile.SetCurrentLayer(y)
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

// NewLayersUI creates the UI representation of the CurrentFile's layers
func NewLayersUI(bounds rl.Rectangle) *Entity {
	// New layer button
	newLayerButton := NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight), "./res/icons/plus.png", false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			CurrentFile.AddNewLayer()
			max := len(CurrentFile.Layers)
			last := CurrentFile.Layers[max-2] // ignore the temp layer

			if currentLayerHoverable != nil {
				currentLayerHoverable.Selected = false
			}

			list.PushChild(LayersUIMakeBox(max-2, last.Name))
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
