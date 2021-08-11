package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	currentLayerHoverable *Hoverable
	layerInteractables    = make(map[int]*Entity)

	layerList          *Entity
	layerListContainer *Entity
)

// LayersUISetCurrentLayer can be used to activate a callback on a layer button
// Intended to be used by the ControlSystem
func LayersUISetCurrentLayer(index int) {
	currentLayerHoverable.Selected = false
	entity, ok := layerInteractables[index]
	if ok {
		if res, err := scene.QueryID(entity.ID); err == nil {
			hoverable := res.Components[entity.Scene.ComponentsMap["hoverable"]].(*Hoverable)
			interactable := res.Components[entity.Scene.ComponentsMap["interactable"]].(*Interactable)

			interactable.OnMouseUp(entity, rl.MouseLeftButton)

			currentLayerHoverable = hoverable
		}
	}
}

// LayersUIMakeList makes the list
func LayersUIMakeList(bounds rl.Rectangle) {
	layerList = NewScrollableList(rl.NewRectangle(0, UIButtonHeight, bounds.Width, bounds.Height-UIButtonHeight), []*Entity{}, FlowDirectionVerticalReversed)
	// All of the layers
	for i, layer := range CurrentFile.Layers {
		if i == len(CurrentFile.Layers)-1 {
			// ignore hidden layer
			continue
		}
		layerList.PushChild(LayersUIMakeBox(i, layer))
	}
	layerList.FlowChildren()
}

// LayersUIRebuildList rebuilds the list
func LayersUIRebuildList() {
	layerList.DestroyNested()
	layerList.Destroy()
	layerListContainer.RemoveChild(layerList)

	if res, err := scene.QueryID(layerListContainer.ID); err == nil {
		moveable := res.Components[layerListContainer.Scene.ComponentsMap["moveable"]].(*Moveable)
		bounds := moveable.Bounds
		LayersUIMakeList(bounds)
		layerListContainer.PushChild(layerList)
		layerListContainer.FlowChildren()
	}
}

// LayersUIMakeBox makes a box
func LayersUIMakeBox(y int, layer *Layer) *Entity {
	var bounds rl.Rectangle
	if res, err := scene.QueryID(layerListContainer.ID); err == nil {
		moveable := res.Components[layerListContainer.Scene.ComponentsMap["moveable"]].(*Moveable)
		bounds = moveable.Bounds
	}

	hidden := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2), GetFile("./res/icons/eye_open.png"), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up

			if res, err := scene.QueryID(entity.ID); err == nil {
				drawable := res.Components[entity.Scene.ComponentsMap["drawable"]].(*Drawable)
				CurrentFile.Layers[y].Hidden = !CurrentFile.Layers[y].Hidden
				drawableTexture, ok := drawable.DrawableType.(*DrawableTexture)
				if ok {
					if CurrentFile.Layers[y].Hidden {
						drawableTexture.SetTexture(GetFile("./res/icons/eye_closed.png"))
					} else {
						drawableTexture.SetTexture(GetFile("./res/icons/eye_open.png"))
					}
				}
			}
		}, nil)
	moveUp := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2), GetFile("./res/icons/arrow_up.png"), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			if err := CurrentFile.MoveLayerUp(y); err == nil {
				if CurrentFile.CurrentLayer == y {
					CurrentFile.SetCurrentLayer(y + 1)
				}
				LayersUIRebuildList()
			}
		}, nil)
	moveDown := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2), GetFile("./res/icons/arrow_down.png"), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			if err := CurrentFile.MoveLayerDown(y); err == nil {
				if CurrentFile.CurrentLayer == y {
					CurrentFile.SetCurrentLayer(y - 1)
				}
				LayersUIRebuildList()
			}
		}, nil)
	delete := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2), GetFile("./res/icons/cross.png"), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			if err := CurrentFile.DeleteLayer(y, true); err == nil {
				LayersUIRebuildList()
			}
		}, nil)

	// Keep the buttons organized
	buttonBox := NewBox(rl.NewRectangle(0, 0, UIButtonHeight*1.5, UIButtonHeight),
		[]*Entity{
			hidden,
			moveUp,
			moveDown,
			delete,
		},
		FlowDirectionHorizontal)

	preview := NewRenderTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight), nil, nil)
	if res, err := scene.QueryID(preview.ID); err == nil {
		drawable := res.Components[preview.Scene.ComponentsMap["drawable"]].(*Drawable)
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			renderTexture.Texture = layer.Canvas
		}
	}

	isCurrent := CurrentFile.CurrentLayer == y
	label := NewInput(rl.NewRectangle(0, 0, bounds.Width-UIButtonHeight*2.5, UIButtonHeight), layer.Name, isCurrent,
		func(entity *Entity, button rl.MouseButton) {
			// button up
		},
		func(entity *Entity, button rl.MouseButton, isHeld bool) {
			if entity == nil {
				// TODO find why the first call is nil
				return
			}
			if hoverable, ok := entity.GetHoverable(); ok {
				if currentLayerHoverable != nil {
					currentLayerHoverable.Selected = false
				}
				currentLayerHoverable = hoverable
				hoverable.Selected = true

				CurrentFile.SetCurrentLayer(y)
			}
		},
		func(entity *Entity, key rl.Key) {
			// key pressed
			if drawable, ok := entity.GetDrawable(); ok {
				if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
					if key == rl.KeyBackspace && len(drawableText.Label) > 0 {
						drawableText.Label = drawableText.Label[:len(drawableText.Label)-1]
					} else if len(drawableText.Label) < 12 {
						switch {
						// 0 to 9
						case key >= 48 && key <= 57:
							fallthrough
						// a to z
						case key >= 97 && key <= 97+26:
							fallthrough
						case key >= rl.KeyA && key <= rl.KeyZ:
							drawableText.Label += string(rune(key))
						}
					}
					CurrentFile.Layers[y].Name = drawableText.Label
				}
			}

		})

	// Set current layer ref
	if res, err := scene.QueryID(label.ID); err == nil {
		hoverable := res.Components[label.Scene.ComponentsMap["hoverable"]].(*Hoverable)

		if isCurrent {
			currentLayerHoverable = hoverable
		}

		layerInteractables[y] = label
	}

	box := NewBox(rl.NewRectangle(0, 0, bounds.Width, UIButtonHeight), []*Entity{
		buttonBox,
		preview,
		label,
	}, FlowDirectionHorizontal)
	return box
}

// NewLayersUI creates the UI representation of the CurrentFile's layers
func NewLayersUI(bounds rl.Rectangle) *Entity {
	// New layer button
	newLayerButton := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight), GetFile("./res/icons/plus.png"), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			CurrentFile.AddNewLayer()
			max := len(CurrentFile.Layers)
			last := CurrentFile.Layers[max-2] // ignore the temp layer

			if currentLayerHoverable != nil {
				currentLayerHoverable.Selected = false
			}

			layerList.PushChild(LayersUIMakeBox(max-2, last))
			layerList.FlowChildren()
		}, nil)

	layerListContainer = NewBox(bounds, []*Entity{
		newLayerButton,
	}, FlowDirectionVertical)

	LayersUIMakeList(bounds)
	layerListContainer.PushChild(layerList)
	layerListContainer.FlowChildren()

	return layerListContainer
}
