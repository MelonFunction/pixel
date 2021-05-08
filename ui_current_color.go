package main

import (
	"log"
	"math"

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

	colorCopy := color
	log.Println(colorCopy)

	if colorCopy.R != colorCopy.G && colorCopy.R != colorCopy.B {
		for colorCopy.R > 0 && colorCopy.G > 0 && colorCopy.B > 0 {
			colorCopy.R--
			colorCopy.G--
			colorCopy.B--
		}

		switch {
		case colorCopy.G == 0 && colorCopy.B == 0:
			colorCopy.R = 255
		case colorCopy.R == 0 && colorCopy.B == 0:
			colorCopy.G = 255
		case colorCopy.R == 0 && colorCopy.G == 0:
			colorCopy.B = 255
		case colorCopy.R == 0:
			log.Println("\tR 0")
			if colorCopy.G > colorCopy.B {
				colorCopy.B = uint8(math.Round(255 / float64(colorCopy.G) * float64(colorCopy.B)))
				colorCopy.G = 255
			} else {
				colorCopy.G = uint8(math.Round(255 / float64(colorCopy.B) * float64(colorCopy.G)))
				colorCopy.B = 255
			}
		case colorCopy.G == 0:
			log.Println("\tG 0")
			if colorCopy.B > colorCopy.R {
				colorCopy.R = uint8(math.Round(255 / float64(colorCopy.B) * float64(colorCopy.R)))
				colorCopy.B = 255
			} else {
				colorCopy.B = uint8(math.Round(255 / float64(colorCopy.R) * float64(colorCopy.B)))
				colorCopy.R = 255
			}
		case colorCopy.B == 0:
			log.Println("\tB 0")
			if colorCopy.R > colorCopy.G {
				colorCopy.G = uint8(math.Round(255 / float64(colorCopy.R) * float64(colorCopy.G)))
				colorCopy.R = 255
			} else {
				colorCopy.R = uint8(math.Round(255 / float64(colorCopy.G) * float64(colorCopy.R)))
				colorCopy.G = 255
			}
		}

	}

	log.Println("\tafter", colorCopy)

	// var scaled float64
	// var rounded float64
	// if color.G > 0 {
	// 	rounded = math.Round(float64((float32(color.G) / float32(color.R)) * 255))
	// 	scaled = float64((float32(color.G) / float32(color.R)) * 255)
	// }

	// TODO
	// Need to search for near values on the value not equal to 255 since the
	// slider increments by 6 or 7 depending on how the value was rounded
	if found, ok := sliderColorsRev[colorCopy]; ok {
		log.Println("\tfound", colorCopy, found)
	} else {
		log.Println("\tnot found, retrying", colorCopy)
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
