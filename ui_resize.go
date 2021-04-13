package main

import (
	"fmt"
	"strconv"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	resizeButtons *Entity
)

func ResizeUIShowDialog() {
	resizeButtons.Show()
}

func NewResizeUI() *Entity {
	var closeResizeButton *Entity

	cx := rl.GetScreenWidth() / 2
	cy := rl.GetScreenHeight() / 2

	bounds := rl.NewRectangle(
		float32(cx)-UIFontSize*15,
		float32(cy)-UIFontSize*5,
		float32(rl.GetScreenWidth()),
		float32(rl.GetScreenHeight()),
	)

	closeResizeButton = NewButtonText(
		rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		"X", false, func(entity *Entity, button rl.MouseButton) {
			resizeButtons.Hide()
		}, nil)
	// closeResizeButton.Hide()

	// Controls for resizing from a particular side
	anchorBox := NewBox(rl.NewRectangle(
		float32(cx),
		float32(cy),
		float32(UIFontSize*2*3),
		float32(UIFontSize*2*3),
	), []*Entity{
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			".", false, func(entity *Entity, button rl.MouseButton) {
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			"^", false, func(entity *Entity, button rl.MouseButton) {
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			".", false, func(entity *Entity, button rl.MouseButton) {
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			"<", false, func(entity *Entity, button rl.MouseButton) {
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			".", false, func(entity *Entity, button rl.MouseButton) {
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			">", false, func(entity *Entity, button rl.MouseButton) {
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			".", false, func(entity *Entity, button rl.MouseButton) {
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			"v", false, func(entity *Entity, button rl.MouseButton) {
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			".", false, func(entity *Entity, button rl.MouseButton) {
			}, nil),
	}, FlowDirectionHorizontal)
	anchorBox.FlowChildren()

	var widthInput, heightInput *Entity
	var newCanvasWidth, newCanvasHeight int

	widthInput = NewInput(rl.NewRectangle(0, 0, UIFontSize*2*10, UIButtonHeight), fmt.Sprint(CurrentFile.CanvasWidth), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
		}, nil,
		func(entity *Entity, key rl.Key) {

			// key pressed
			if res, err := scene.QueryID(entity.ID); err == nil {
				drawable := res.Components[entity.Scene.ComponentsMap["drawable"]].(*Drawable)
				drawableParent, ok := drawable.DrawableType.(*DrawableText)

				if ok {
					finishedEditing := func() {
						if parsed, err := strconv.ParseInt(drawableParent.Label, 10, 64); err == nil {
							newCanvasWidth = int(parsed)
							// TODO call layer resize function
						}
						RemoveCapturedInput()
					}

					switch {
					case key >= 48 && key <= 57:
						drawableParent.Label += string(rune(key))
					case key == rl.KeyBackspace:
						drawableParent.Label = drawableParent.Label[:len(drawableParent.Label)-1]
					case key == rl.KeyTab:
						finishedEditing()
						// Set control to heightInput
						if hi, ok := heightInput.GetInteractable(); ok {
							UIEntityCapturedInput = heightInput
							UIInteractableCapturedInput = hi
						}
					case key == rl.KeyEnter:
						finishedEditing()
					}
				}
			}

		})

	heightInput = NewInput(rl.NewRectangle(0, 0, UIFontSize*2*10, UIButtonHeight), fmt.Sprint(CurrentFile.CanvasHeight), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
		}, nil,
		func(entity *Entity, key rl.Key) {

			// key pressed
			if res, err := scene.QueryID(entity.ID); err == nil {
				drawable := res.Components[entity.Scene.ComponentsMap["drawable"]].(*Drawable)
				drawableParent, ok := drawable.DrawableType.(*DrawableText)

				if ok {
					finishedEditing := func() {
						if parsed, err := strconv.ParseInt(drawableParent.Label, 10, 64); err == nil {
							newCanvasHeight = int(parsed)
							// TODO call layer resize function
						}
						RemoveCapturedInput()
					}

					switch {
					case key >= 48 && key <= 57:
						drawableParent.Label += string(rune(key))
					case key == rl.KeyBackspace:
						drawableParent.Label = drawableParent.Label[:len(drawableParent.Label)-1]
					case key == rl.KeyTab:
						finishedEditing()
					case key == rl.KeyEnter:
						// TODO make this do resize event and then close resize dialogue
						finishedEditing()
					}
				}
			}

		})

	textBoxes := NewBox(rl.NewRectangle(
		float32(cx),
		float32(cy),
		float32(UIFontSize*2*10),
		float32(UIFontSize*2*10),
	), []*Entity{
		widthInput,
		heightInput,
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2*10, UIButtonHeight),
			"Resize", false, func(entity *Entity, button rl.MouseButton) {
			}, nil),
	}, FlowDirectionVertical)

	// Added to scene on first hover
	resizeButtons = NewBox(
		bounds,
		[]*Entity{
			closeResizeButton,
			anchorBox,
			textBoxes,
		},
		FlowDirectionHorizontal,
	)
	resizeButtons.FlowChildren()

	return resizeButtons
}
