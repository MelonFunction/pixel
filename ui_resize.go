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
	CurrentFile.DoingResize = true
}

func ResizeUIHideDialog() {
	resizeButtons.Hide()
	CurrentFile.DoingResize = false
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
			ResizeUIHideDialog()
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
				CurrentFile.CanvasDirectionResizePreview = ResizeTL
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			"^", false, func(entity *Entity, button rl.MouseButton) {
				CurrentFile.CanvasDirectionResizePreview = ResizeTC
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			".", false, func(entity *Entity, button rl.MouseButton) {
				CurrentFile.CanvasDirectionResizePreview = ResizeTR
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			"<", false, func(entity *Entity, button rl.MouseButton) {
				CurrentFile.CanvasDirectionResizePreview = ResizeCL
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			".", false, func(entity *Entity, button rl.MouseButton) {
				CurrentFile.CanvasDirectionResizePreview = ResizeCC
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			">", false, func(entity *Entity, button rl.MouseButton) {
				CurrentFile.CanvasDirectionResizePreview = ResizeCR
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			".", false, func(entity *Entity, button rl.MouseButton) {
				CurrentFile.CanvasDirectionResizePreview = ResizeBL
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			"v", false, func(entity *Entity, button rl.MouseButton) {
				CurrentFile.CanvasDirectionResizePreview = ResizeBC
			}, nil),
		NewButtonText(
			rl.NewRectangle(0, 0, UIFontSize*2, UIFontSize*2),
			".", false, func(entity *Entity, button rl.MouseButton) {
				CurrentFile.CanvasDirectionResizePreview = ResizeBR
			}, nil),
	}, FlowDirectionHorizontal)
	anchorBox.FlowChildren()

	makeInput := func(linkedValue *int, tabNext *Entity) *Entity {
		return NewInput(rl.NewRectangle(0, 0, UIFontSize*2*10, UIButtonHeight), fmt.Sprint(*linkedValue), false,
			func(entity *Entity, button rl.MouseButton) {
				// button up
			}, nil,
			func(entity *Entity, key rl.Key) {
				// key pressed
				if res, err := scene.QueryID(entity.ID); err == nil {
					drawable := res.Components[entity.Scene.ComponentsMap["drawable"]].(*Drawable)
					drawableParent, ok := drawable.DrawableType.(*DrawableText)

					if ok {
						alterValue := func() {
							if parsed, err := strconv.ParseInt(drawableParent.Label, 10, 64); err == nil {
								*linkedValue = int(parsed)
							}
						}

						switch {
						case key >= 48 && key <= 57:
							drawableParent.Label += string(rune(key))
							alterValue()
						case key == rl.KeyBackspace && len(drawableParent.Label) > 0:
							drawableParent.Label = drawableParent.Label[:len(drawableParent.Label)-1]
							alterValue()
						case key == rl.KeyTab:
							RemoveCapturedInput()

							// Set control to tabNext
							if tabNext != nil {
								if hi, ok := tabNext.GetInteractable(); ok {
									UIEntityCapturedInput = tabNext
									UIInteractableCapturedInput = hi
								}
							}
						case key == rl.KeyEnter:
							RemoveCapturedInput()
						}
					}
				}
			})
	}

	heightInput := makeInput(&CurrentFile.CanvasHeightResizePreview, nil)
	widthInput := makeInput(&CurrentFile.CanvasWidthResizePreview, heightInput)

	canvasTextBoxes := NewBox(rl.NewRectangle(
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
				CurrentFile.Resize(CurrentFile.CanvasWidthResizePreview, CurrentFile.CanvasHeightResizePreview, CurrentFile.CanvasDirectionResizePreview)
			}, nil),
	}, FlowDirectionVertical)

	// Added to scene on first hover
	resizeButtons = NewBox(
		bounds,
		[]*Entity{
			closeResizeButton,
			anchorBox,
			canvasTextBoxes,
		},
		FlowDirectionHorizontal,
	)
	resizeButtons.FlowChildren()

	ResizeUIShowDialog()

	return resizeButtons
}
