package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	currentToolHoverable *Hoverable
	toolsButtons         *Entity
	toolPencil           *Entity
	toolEraser           *Entity
	toolFill             *Entity
	toolPicker           *Entity
	toolSelector         *Entity
)

func ToolsUISetCurrentToolSelected(entity *Entity) {
	if hoverable, ok := entity.GetHoverable(); ok {
		if currentToolHoverable != nil {
			currentToolHoverable.Selected = false
		}
		currentToolHoverable = hoverable
		hoverable.Selected = true
	}
}

// NewToolsUI creates and returns the tools UI entity
func NewToolsUI(bounds rl.Rectangle) *Entity {
	toolsButtons = NewBox(bounds, []*Entity{}, FlowDirectionHorizontal)

	// TODO allow right click to be replaced with selector if alt is pressed
	toolPencil = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		"./res/icons/pencil.png", false, func(entity *Entity, button rl.MouseButton) {
			CurrentFile.LeftTool = NewPixelBrushTool("Pixel Brush", false)
			CurrentFile.RightTool = NewPixelBrushTool("Pixel Brush", false)
			ToolsUISetCurrentToolSelected(entity)
		}, nil)
	toolEraser = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		"./res/icons/eraser.png", false, func(entity *Entity, button rl.MouseButton) {
			CurrentFile.LeftTool = NewPixelBrushTool("Eraser", true)
			CurrentFile.RightTool = NewPixelBrushTool("Eraser", true)
			ToolsUISetCurrentToolSelected(entity)
		}, nil)
	toolFill = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		"./res/icons/fill.png", false, func(entity *Entity, button rl.MouseButton) {
			CurrentFile.LeftTool = NewFillTool("Fill")
			CurrentFile.RightTool = NewFillTool("Fill")
			ToolsUISetCurrentToolSelected(entity)
		}, nil)
	toolPicker = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		"./res/icons/picker.png", false, func(entity *Entity, button rl.MouseButton) {
			CurrentFile.LeftTool = NewPickerTool("Picker")
			CurrentFile.RightTool = NewPickerTool("Picker")
			ToolsUISetCurrentToolSelected(entity)
		}, nil)
	toolSelector = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		"./res/icons/selector.png", false, func(entity *Entity, button rl.MouseButton) {
			CurrentFile.LeftTool = NewSelectorTool("Selector")
			CurrentFile.RightTool = NewSelectorTool("Selector")
			ToolsUISetCurrentToolSelected(entity)
		}, nil)

	toolsButtons.PushChild(toolPencil)
	toolsButtons.PushChild(toolEraser)
	toolsButtons.PushChild(toolFill)
	toolsButtons.PushChild(toolPicker)
	toolsButtons.PushChild(toolSelector)
	toolsButtons.FlowChildren()

	ToolsUISetCurrentToolSelected(toolPencil)

	return toolsButtons
}
