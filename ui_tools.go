package main

import (
	"fmt"
	"strconv"

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
	toolSettings         *Entity // extra space which can be used by other ui
)

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
	case toolPencil:
		var size int
		if lt, ok := CurrentFile.LeftTool.(*PixelBrushTool); ok {
			size = lt.GetSize()
		}
		brushWidthInput := NewInput(rl.NewRectangle(0, 0, UIButtonHeight*3, UIButtonHeight), fmt.Sprintf("%d", size), false,
			func(entity *Entity, button rl.MouseButton) {
				// button up
			},
			nil,
			func(entity *Entity, key rl.Key) {
				// key pressed
				if drawable, ok := entity.GetDrawable(); ok {
					if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
						// TODO this could probably be added to util since the same
						// code exists in multiple places
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

							if i, err := strconv.ParseInt(drawableText.Label, 10, 64); err == nil {
								// Set tools from label
								if lt, ok := CurrentFile.LeftTool.(*PixelBrushTool); ok {
									lt.SetSize(int(i))

									// Set label text
									drawableText.Label = fmt.Sprintf("%d", lt.GetSize())
								}
								if rt, ok := CurrentFile.RightTool.(*PixelBrushTool); ok {
									rt.SetSize(int(i))
								}

							}

						}
					}
				}
			})
		if interactable, ok := brushWidthInput.GetInteractable(); ok {
			interactable.OnScroll = func(direction int) {
				if drawable, ok := brushWidthInput.GetDrawable(); ok {
					if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
						if lt, ok := CurrentFile.LeftTool.(*PixelBrushTool); ok {
							lt.SetSize(lt.GetSize() + direction)
							drawableText.Label = fmt.Sprintf("%d", lt.GetSize())
						}
						if rt, ok := CurrentFile.RightTool.(*PixelBrushTool); ok {
							rt.SetSize(rt.GetSize() + direction)
							drawableText.Label = fmt.Sprintf("%d", rt.GetSize())
						}
					}
				}
			}
		}
		toolSettings.PushChild(brushWidthInput)
	}

	toolSettings.FlowChildren()
}

// NewToolsUI creates and returns the tools UI entity
func NewToolsUI(bounds rl.Rectangle) *Entity {
	toolsButtons = NewBox(bounds, []*Entity{}, FlowDirectionHorizontal)

	// TODO allow right click to be replaced with selector if alt is pressed
	toolPencil = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/pencil.png"), false, func(entity *Entity, button rl.MouseButton) {
			// Commit the selection, stop showing selection preview etc
			if len(CurrentFile.Selection) > 0 {
				CurrentFile.CommitSelection()
			}
			CurrentFile.LeftTool = NewPixelBrushTool("Pixel Brush", false)
			CurrentFile.RightTool = NewPixelBrushTool("Pixel Brush", false)
			ToolsUISetCurrentToolSelected(entity)
		}, nil)
	toolEraser = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/eraser.png"), false, func(entity *Entity, button rl.MouseButton) {
			// Commit the selection, stop showing selection preview etc
			if len(CurrentFile.Selection) > 0 {
				CurrentFile.CommitSelection()
			}
			CurrentFile.LeftTool = NewPixelBrushTool("Eraser", true)
			CurrentFile.RightTool = NewPixelBrushTool("Eraser", true)
			ToolsUISetCurrentToolSelected(entity)
		}, nil)
	toolFill = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/fill.png"), false, func(entity *Entity, button rl.MouseButton) {
			// Commit the selection, stop showing selection preview etc
			if len(CurrentFile.Selection) > 0 {
				CurrentFile.CommitSelection()
			}
			CurrentFile.LeftTool = NewFillTool("Fill")
			CurrentFile.RightTool = NewFillTool("Fill")
			ToolsUISetCurrentToolSelected(entity)
		}, nil)
	toolPicker = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/picker.png"), false, func(entity *Entity, button rl.MouseButton) {
			// Commit the selection, stop showing selection preview etc
			if len(CurrentFile.Selection) > 0 {
				CurrentFile.CommitSelection()
			}
			CurrentFile.LeftTool = NewPickerTool("Picker")
			CurrentFile.RightTool = NewPickerTool("Picker")
			ToolsUISetCurrentToolSelected(entity)
		}, nil)
	toolSelector = NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		GetFile("./res/icons/selector.png"), false, func(entity *Entity, button rl.MouseButton) {
			CurrentFile.LeftTool = NewSelectorTool("Selector")
			CurrentFile.RightTool = NewSelectorTool("Selector")
			ToolsUISetCurrentToolSelected(entity)
		}, nil)

	// currently only 5 buttons
	bounds.Width -= UIButtonHeight * 5
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
