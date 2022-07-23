package main

import (
	"fmt"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	currentToolHoverable *Hoverable
	toolsButtons         *Entity
	toolPencil           *Entity
	toolEraser           *Entity
	toolFill             *Entity
	toolPicker           *Entity
	toolSelector         *Entity
	toolSettings         *Entity // extra space which can be used by other ui
)

// ToolsUISetCurrentToolSelected makes the tool have the selected appearance
// It also changes the UI to show additional items in the empty space to the
// right of the tools
func ToolsUISetCurrentToolSelected(entity *Entity) {
	if hoverable, ok := entity.GetHoverable(); ok {
		if currentToolHoverable != nil {
			currentToolHoverable.Selected = false
		}
		currentToolHoverable = hoverable
		hoverable.Selected = true
	}

	toolSettings.RemoveChildren()

	switch entity {
	case toolEraser:
		fallthrough
	case toolPencil:
		var size int32
		var shape BrushShape
		if lt, ok := LeftTool.(*PixelBrushTool); ok {
			size = lt.GetSize()
			shape = lt.GetShape()
		}
		brushShapeBox := NewBox(rl.NewRectangle(0, 0, UIButtonHeight*0.5, UIButtonHeight), []*Entity{
			NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2), GetFile("./res/icons/circle.png"), shape == BrushShapeCircle,
				func(e *Entity, button MouseButton) {
					// button up
					if lt, ok := LeftTool.(*PixelBrushTool); ok {
						lt.SetShape(BrushShapeCircle)
					}
					if rt, ok := RightTool.(*PixelBrushTool); ok {
						rt.SetShape(BrushShapeCircle)
					}
					ToolsUISetCurrentToolSelected(entity)
				}, nil),
			NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2), GetFile("./res/icons/square.png"), shape == BrushShapeSquare,
				func(e *Entity, button MouseButton) {
					// button up
					if lt, ok := LeftTool.(*PixelBrushTool); ok {
						lt.SetShape(BrushShapeSquare)
					}
					if rt, ok := RightTool.(*PixelBrushTool); ok {
						rt.SetShape(BrushShapeSquare)
					}
					ToolsUISetCurrentToolSelected(entity)
				}, nil),
		}, FlowDirectionVertical)
		brushWidthInput := NewInput(rl.NewRectangle(0, 0, UIButtonHeight*3, UIButtonHeight), fmt.Sprintf("%d", size), TextAlignCenter, false,
			func(entity *Entity, button MouseButton) {
				// button up
			},
			nil,
			func(entity *Entity, key Key) {
				// key pressed
				if drawable, ok := entity.GetDrawable(); ok {
					if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
						// TODO this could probably be added to util since the same
						// code exists in multiple places
						if key == rl.KeyBackspace && len(drawableText.Label) > 0 {
							drawableText.Label = drawableText.Label[:len(drawableText.Label)-1]
						} else if len(drawableText.Label) < 12 {
							if key >= 48 && key <= 57 { // 0 to 9
								drawableText.Label += string(rune(key))
							}

							if i, err := strconv.ParseInt(drawableText.Label, 10, 64); err == nil {
								// Set tools from label
								if lt, ok := LeftTool.(*PixelBrushTool); ok {
									lt.SetSize(int32(i))

									// Set label text
									drawableText.Label = fmt.Sprintf("%d", lt.GetSize())
								}
								if rt, ok := RightTool.(*PixelBrushTool); ok {

									rt.SetSize(int32(i))
								}

							}

						}
					}
				}
			})
		if interactable, ok := brushWidthInput.GetInteractable(); ok {
			interactable.OnScroll = func(direction int32) {
				if drawable, ok := brushWidthInput.GetDrawable(); ok {
					if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
						if lt, ok := LeftTool.(*PixelBrushTool); ok {
							lt.SetSize(lt.GetSize() + direction)
							drawableText.Label = fmt.Sprintf("%d", lt.GetSize())
						}
						if rt, ok := RightTool.(*PixelBrushTool); ok {
							rt.SetSize(rt.GetSize())
							drawableText.Label = fmt.Sprintf("%d", rt.GetSize())
						}
					}
				}
			}
		}
		toolSettings.PushChild(brushShapeBox)
		toolSettings.PushChild(brushWidthInput)
	}

	toolSettings.FlowChildren()
}

// NewToolsUI creates and returns the tools UI entity
func NewToolsUI(bounds rl.Rectangle) *Entity {
	toolsButtons = NewBox(bounds, []*Entity{}, FlowDirectionHorizontal)

	// TODO allow right click to be replaced with selector if alt is pressed
	toolPencil = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/pencil.png"), false, func(entity *Entity, button MouseButton) {
			// Commit the selection, stop showing selection preview etc
			if len(CurrentFile.Selection) > 0 {
				CurrentFile.CommitSelection()
			}
			LeftTool = NewPixelBrushTool("Pixel Brush", false)
			RightTool = NewPixelBrushTool("Pixel Brush", false)
			ToolsUISetCurrentToolSelected(entity)
		}, nil)
	toolEraser = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/eraser.png"), false, func(entity *Entity, button MouseButton) {
			// Commit the selection, stop showing selection preview etc
			if len(CurrentFile.Selection) > 0 {
				CurrentFile.CommitSelection()
			}
			LeftTool = NewPixelBrushTool("Eraser", true)
			RightTool = NewPixelBrushTool("Eraser", true)
			ToolsUISetCurrentToolSelected(entity)
		}, nil)
	toolFill = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/fill.png"), false, func(entity *Entity, button MouseButton) {
			// Commit the selection, stop showing selection preview etc
			if len(CurrentFile.Selection) > 0 {
				CurrentFile.CommitSelection()
			}
			LeftTool = NewFillTool("Fill")
			RightTool = NewFillTool("Fill")
			ToolsUISetCurrentToolSelected(entity)
		}, nil)
	toolPicker = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/picker.png"), false, func(entity *Entity, button MouseButton) {
			// Commit the selection, stop showing selection preview etc
			if len(CurrentFile.Selection) > 0 {
				CurrentFile.CommitSelection()
			}
			LeftTool = NewPickerTool("Picker")
			RightTool = NewPickerTool("Picker")
			ToolsUISetCurrentToolSelected(entity)
		}, nil)
	toolSelector = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/selector.png"), false, func(entity *Entity, button MouseButton) {
			LeftTool = NewSelectorTool("Selector")
			RightTool = NewSelectorTool("Selector")
			ToolsUISetCurrentToolSelected(entity)
		}, nil)

	// currently only 5 buttons
	// bounds.Width = UIButtonHeight
	toolSettings = NewBox(bounds, []*Entity{}, FlowDirectionHorizontal)

	toolsButtons.PushChild(toolPencil)
	toolsButtons.PushChild(toolEraser)
	toolsButtons.PushChild(toolFill)
	toolsButtons.PushChild(toolPicker)
	toolsButtons.PushChild(toolSelector)
	toolsButtons.PushChild(toolSettings)
	toolsButtons.FlowChildren()

	ToolsUISetCurrentToolSelected(toolPencil)

	return toolsButtons
}
