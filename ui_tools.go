package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	toolsButtons *Entity
)

func ToolsUICloseEditor() {

}

func ToolsUIAddButton() {
	pencil := NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight),
		"./res/icons/pencil.png", true, func(entity *Entity, button rl.MouseButton) {
			switch button {
			case rl.MouseLeftButton:
				CurrentFile.LeftTool = NewPixelBrushTool("Pixel Brush", false)
			case rl.MouseRightButton:
				CurrentFile.RightTool = NewPixelBrushTool("Pixel Brush", false)
			}
		}, nil)
	eraser := NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight),
		"./res/icons/eraser.png", true, func(entity *Entity, button rl.MouseButton) {
			switch button {
			case rl.MouseLeftButton:
				CurrentFile.LeftTool = NewPixelBrushTool("Eraser", true)
			case rl.MouseRightButton:
				CurrentFile.RightTool = NewPixelBrushTool("Eraser", true)
			}
		}, nil)

	toolsButtons.PushChild(pencil)
	toolsButtons.PushChild(eraser)
	toolsButtons.FlowChildren()
}

func NewToolsUI(bounds rl.Rectangle) *Entity {
	toolsButtons = NewBox(bounds, []*Entity{}, FlowDirectionVertical)
	ToolsUIAddButton()
	return toolsButtons
}
