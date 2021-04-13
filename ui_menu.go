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

type MenuContextOptions struct {
	Entity  *Entity
	Visible bool
}

func NewMenuUI(bounds rl.Rectangle) *Entity {
	menuButtons = NewBox(bounds, []*Entity{}, FlowDirectionHorizontal)
	var saveButton, exportButton, openButton, resizeButton, fileButton *Entity
	hoveredButtons := make([]*Entity, 0, 4)

	// Show all of the buttons when file button or its contents are hovered
	showFileButtons := func(entity *Entity) func() {
		return func() {
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
				saveButton.Show()
				exportButton.Show()
				openButton.Show()
				resizeButton.Show()
				menuContexts.Scene.MoveEntityToEnd(menuContexts)
			}
		}
	}

	hideFileButtons := func(entity *Entity) func() {
		return func() {
			for i, e := range hoveredButtons {
				if e == entity {
					hoveredButtons = append(hoveredButtons[:i], hoveredButtons[i+1:]...)
				}
			}

			// Hide everything if nothing is being hovered
			go func() {
				time.Sleep(500 * time.Millisecond)
				if len(hoveredButtons) == 0 {
					saveButton.Hide()
					exportButton.Hide()
					openButton.Hide()
					resizeButton.Hide()
				}
			}()
		}
	}

	fo := rl.MeasureTextEx(*Font, "resize", UIFontSize, 1)
	saveButton = NewButtonText(
		rl.NewRectangle(0, UIFontSize*2, fo.X+10, UIFontSize*2),
		"save", false, func(entity *Entity, button rl.MouseButton) {
			UISave()
		}, nil)
	saveButton.Hide()

	exportButton = NewButtonText(
		rl.NewRectangle(0, UIFontSize*4, fo.X+10, UIFontSize*2),
		"export", false, func(entity *Entity, button rl.MouseButton) {
			UIExport()
		}, nil)
	exportButton.Hide()

	openButton = NewButtonText(
		rl.NewRectangle(0, UIFontSize*6, fo.X+10, UIFontSize*2),
		"open", false, func(entity *Entity, button rl.MouseButton) {
			UIOpen()
		}, nil)
	openButton.Hide()

	resizeButton = NewButtonText(
		rl.NewRectangle(0, UIFontSize*8, fo.X+10, UIFontSize*2),
		"resize", false, func(entity *Entity, button rl.MouseButton) {
			UIExport()
		}, nil)
	resizeButton.Hide()

	// "Parent" button
	fo = rl.MeasureTextEx(*Font, "file", UIFontSize, 1)
	fileButton = NewButtonText(
		rl.NewRectangle(0, 0, fo.X+10, UIFontSize*2),
		"file", false, func(entity *Entity, button rl.MouseButton) {
		}, nil)
	menuButtons.PushChild(fileButton)

	for _, button := range []*Entity{saveButton, exportButton, openButton, resizeButton, fileButton} {
		if hoverable, ok := button.GetHoverable(); ok {
			hoverable.OnMouseEnter = showFileButtons(button)
			hoverable.OnMouseLeave = hideFileButtons(button)
		}
	}

	// Added to scene on first hover
	menuContexts = NewBox(bounds, []*Entity{
		saveButton,
		exportButton,
		openButton,
		resizeButton,
	}, FlowDirectionNone)

	menuButtons.FlowChildren()
	return menuButtons
}
