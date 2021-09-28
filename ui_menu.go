package main

import (
	"time"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	// the buttons themselves
	menuButtons *Entity
	// the dropdown menu
	menuContexts *Entity
)

// NewMenuUI returns a new entity
func NewMenuUI(bounds rl.Rectangle) *Entity {
	var newButton, saveButton, saveAsButton, openButton, resizeButton, fileButton *Entity
	hoveredButtons := make([]*Entity, 0, 4)

	measured := rl.MeasureTextEx(*Font, "  save as  ", UIFontSize, 1)

	newButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"new", false, func(entity *Entity, button rl.MouseButton) {
			UINew()
		}, nil)
	newButton.Hide()

	saveButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"save", false, func(entity *Entity, button rl.MouseButton) {
			if len(CurrentFile.FileDir) > 0 {
				CurrentFile.SaveAs(CurrentFile.FileDir)
			} else {
				UISaveAs()
			}
		}, nil)
	saveButton.Hide()

	saveAsButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"save as", false, func(entity *Entity, button rl.MouseButton) {
			UISaveAs()
		}, nil)
	saveAsButton.Hide()

	openButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"open", false, func(entity *Entity, button rl.MouseButton) {
			UIOpen()
		}, nil)
	openButton.Hide()

	resizeButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"resize", false, func(entity *Entity, button rl.MouseButton) {
			ResizeUIShowDialog()
		}, nil)
	resizeButton.Hide()

	// "Parent" button
	measured = rl.MeasureTextEx(*Font, "file", UIFontSize, 1)

	fileButton = NewButtonText(
		rl.NewRectangle(100, 100, measured.X+10, UIFontSize*2),
		"file", false, func(entity *Entity, button rl.MouseButton) {
		}, nil)
	menuButtons = NewBox(bounds, []*Entity{
		fileButton,
	}, FlowDirectionHorizontal)

	menuButtons.FlowChildren()

	// Added to scene on first hover
	bounds.Y += UIFontSize * 2
	bounds.Height = float32(rl.GetScreenHeight())
	menuContexts = NewBox(bounds, []*Entity{
		newButton,
		saveButton,
		saveAsButton,
		openButton,
		resizeButton,
	}, FlowDirectionVertical)

	menuContexts.FlowChildren()
	menuContexts.Hide()

	for _, button := range []*Entity{newButton, saveButton, saveAsButton, openButton, resizeButton, fileButton} {
		if hoverable, ok := button.GetHoverable(); ok {
			hoverable.OnMouseEnter = func(entity *Entity) {
				found := false
				for _, e := range hoveredButtons {
					if e == entity {
						found = true
					}
				}
				if !found {
					hoveredButtons = append(hoveredButtons, entity)
				}

				if len(hoveredButtons) > 0 {
					newButton.Show()
					saveButton.Show()
					saveAsButton.Show()
					openButton.Show()
					resizeButton.Show()
					menuContexts.Show()
					menuContexts.Scene.MoveEntityToEnd(menuContexts)
				}
			}
			hoverable.OnMouseLeave = func(entity *Entity) {
				for i, e := range hoveredButtons {
					if e == entity {
						hoveredButtons = append(hoveredButtons[:i], hoveredButtons[i+1:]...)
					}
				}

				// Hide everything if nothing is being hovered
				go func() {
					time.Sleep(500 * time.Millisecond)
					if len(hoveredButtons) == 0 {
						newButton.Hide()
						saveButton.Hide()
						saveAsButton.Hide()
						openButton.Hide()
						resizeButton.Hide()
						menuContexts.Hide()
					}
				}()
			}
		}
	}

	return menuButtons
}
