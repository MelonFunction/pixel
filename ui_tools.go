package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	toolsButtons *Entity
)

// NewToolsUI creates and returns the tools UI entity
func NewToolsUI(bounds rl.Rectangle) *Entity {
	toolsButtons = NewBox(bounds, []*Entity{}, FlowDirectionHorizontal)

	var currentToolHoverable *Hoverable

	setCurrentTool := func(entity *Entity) {
		if hoverable, ok := entity.GetHoverable(); ok {
			hoverable.Selected = true
			if currentToolHoverable != nil {
				currentToolHoverable.Selected = false
			}
			currentToolHoverable = hoverable
		}
	}

	// TODO allow right click to be replaced with selector if alt is pressed
	pencil := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		"./res/icons/pencil.png", false, func(entity *Entity, button rl.MouseButton) {
			CurrentFile.LeftTool = NewPixelBrushTool("Pixel Brush", false)
			CurrentFile.RightTool = NewPixelBrushTool("Pixel Brush", false)
			setCurrentTool(entity)
		}, nil)
	eraser := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		"./res/icons/eraser.png", false, func(entity *Entity, button rl.MouseButton) {
			CurrentFile.LeftTool = NewPixelBrushTool("Eraser", true)
			CurrentFile.RightTool = NewPixelBrushTool("Eraser", true)
			setCurrentTool(entity)
		}, nil)
	fill := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		"./res/icons/fill.png", false, func(entity *Entity, button rl.MouseButton) {
			CurrentFile.LeftTool = NewFillTool("Fill")
			CurrentFile.RightTool = NewFillTool("Fill")
			setCurrentTool(entity)
		}, nil)
	picker := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		"./res/icons/picker.png", false, func(entity *Entity, button rl.MouseButton) {
			CurrentFile.LeftTool = NewPickerTool("Picker")
			CurrentFile.RightTool = NewPickerTool("Picker")
			setCurrentTool(entity)
		}, nil)

	toolsButtons.PushChild(pencil)
	toolsButtons.PushChild(eraser)
	toolsButtons.PushChild(fill)
	toolsButtons.PushChild(picker)
	toolsButtons.FlowChildren()

	setCurrentTool(pencil)

	return toolsButtons
}
