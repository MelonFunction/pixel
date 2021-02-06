package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	toolsButtons *Entity
	currentTool  *Entity
)

func ToolsUICloseEditor() {

}

func ToolsUIAddButton() {
	pencil := NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight),
		"./res/icons/pencil.png", true, func(entity *Entity, button rl.MouseButton) {
			log.Println(button)
			switch button {
			case rl.MouseLeftButton:
				CurrentFile.LeftTool = NewPixelBrushTool("Pixel Brush")
			case rl.MouseRightButton:
				CurrentFile.RightTool = NewPixelBrushTool("Pixel Brush")
			}
		}, nil)
	eraser := NewButtonTexture(rl.NewRectangle(0, 0, buttonHeight, buttonHeight),
		"./res/icons/eraser.png", true, func(entity *Entity, button rl.MouseButton) {
			switch button {
			case rl.MouseLeftButton:
				CurrentFile.LeftTool = NewPixelBrushTool("Eraser")
				CurrentFile.LeftColor = rl.Transparent
			case rl.MouseRightButton:
				CurrentFile.RightTool = NewPixelBrushTool("Eraser")
				CurrentFile.RightColor = rl.Transparent
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
