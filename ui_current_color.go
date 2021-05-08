package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	// The container
	currentColorBox *Entity

	currentColorLeft  *Entity
	currentColorRight *Entity
	currentColorSwap  *Entity
	currentColorAdd   *Entity
)

// CurrentColorSetLeftColor sets the left color and updates the UI components
// to reflect the set color
func CurrentColorSetLeftColor(color rl.Color) {
	if found, ok := areaColorsRev[color]; ok {
		log.Println(color, found)
	} else {
		log.Println("not found", color)
	}

	if drawable, ok := currentColorLeft.GetDrawable(); ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			CurrentFile.LeftColor = color

			texture := renderTexture.Texture
			rl.BeginTextureMode(texture)
			rl.ClearBackground(color)
			rl.EndTextureMode()
		}
	}
}

// CurrentColorSetRightColor sets the right color and updates the UI components
// to reflect the set color
func CurrentColorSetRightColor(color rl.Color) {
	if drawable, ok := currentColorRight.GetDrawable(); ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			CurrentFile.RightColor = color

			texture := renderTexture.Texture
			rl.BeginTextureMode(texture)
			rl.ClearBackground(color)
			rl.EndTextureMode()
		}
	}
}

func currentColorUIAddColor(color rl.Color) *Entity {
	var w float32
	var h float32
	if res, err := scene.QueryID(currentColorBox.ID); err == nil {
		moveable := res.Components[currentColorBox.Scene.ComponentsMap["moveable"]].(*Moveable)
		w = moveable.Bounds.Width / 4
		h = moveable.Bounds.Width / 4
	}

	e := NewRenderTexture(rl.NewRectangle(0, 0, w, h), nil, nil)

	currentColorBox.PushChild(e)
	currentColorBox.FlowChildren()
	return e
}

// NewCurrentColorUI creates a new Current Color UI component which displays
// the currently selected colors as well as buttons to swap them and add the
// color to the palette depending on which mouse button was clicked
func NewCurrentColorUI(bounds rl.Rectangle) *Entity {
	currentColorBox = NewBox(bounds, []*Entity{}, FlowDirectionHorizontal)

	currentColorLeft = currentColorUIAddColor(CurrentFile.LeftColor)
	CurrentColorSetLeftColor(CurrentFile.LeftColor)
	currentColorRight = currentColorUIAddColor(CurrentFile.RightColor)
	CurrentColorSetRightColor(CurrentFile.RightColor)

	currentColorSwap = NewButtonTexture(rl.NewRectangle(0, 0, bounds.Width/4, bounds.Width/4), "./res/icons/swap.png", false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			left := CurrentFile.LeftColor
			right := CurrentFile.RightColor
			CurrentColorSetLeftColor(right)
			CurrentColorSetRightColor(left)
		}, nil)
	currentColorBox.PushChild(currentColorSwap)

	currentColorAdd = NewButtonTexture(rl.NewRectangle(0, 0, bounds.Width/4, bounds.Width/4), "./res/icons/plus.png", false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			switch button {
			case rl.MouseLeftButton:
				PaletteUIAddColor(CurrentFile.LeftColor)
			case rl.MouseRightButton:
				PaletteUIAddColor(CurrentFile.RightColor)
			}
		}, nil)
	currentColorBox.PushChild(currentColorAdd)
	currentColorBox.FlowChildren()

	return currentColorBox
}
