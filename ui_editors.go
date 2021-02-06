package main

import rl "github.com/lachee/raylib-goplus/raylib"

var (
	editorsButtons *Entity
	currentButton  *Entity
)

func EditorsUICloseEditor() {

}

func EditorsUIAddButton(file *File) {
	isCurrent := file == CurrentFile

	fo := rl.MeasureTextEx(*Font, file.Filename, UIFontSize, 1)
	button := NewButtonText(
		rl.NewRectangle(0, 0, fo.X+10, UIFontSize*2),
		file.Filename, isCurrent, func(entity *Entity, button rl.MouseButton) {

			if res, err := scene.QueryID(currentButton.ID); err == nil {
				hoverable := res.Components[currentButton.Scene.ComponentsMap["hoverable"]].(*Hoverable)
				hoverable.Selected = false
			}

			if res, err := scene.QueryID(entity.ID); err == nil {
				hoverable := res.Components[entity.Scene.ComponentsMap["hoverable"]].(*Hoverable)
				hoverable.Selected = true
			}

			CurrentFile = file
			currentButton = entity

			LayersUIRebuildList()
		}, nil)
	if isCurrent {
		// deselect old currentButton
		if currentButton != nil {
			if res, err := scene.QueryID(currentButton.ID); err == nil {
				hoverable := res.Components[currentButton.Scene.ComponentsMap["hoverable"]].(*Hoverable)
				hoverable.Selected = false
			}
			LayersUIRebuildList()
		}

		currentButton = button
	}

	editorsButtons.PushChild(button)
	editorsButtons.FlowChildren()
}

func NewEditorsUI(bounds rl.Rectangle) *Entity {
	editorsButtons = NewScrollableList(bounds, []*Entity{}, FlowDirectionHorizontal)
	for _, f := range Files {
		EditorsUIAddButton(f)
	}
	return editorsButtons
}
