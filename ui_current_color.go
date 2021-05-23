package main

import (
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

type colorEditing int

const (
	editingR colorEditing = iota
	editingG
	editingB
)

func SetUIColors(color rl.Color) {
	cc := color
	// If it's R, G or B which is incremented in the loop
	var editing colorEditing

	for cc.R > 0 && cc.G > 0 && cc.B > 0 {
		cc.R--
		cc.G--
		cc.B--
	}

	switch {
	// For cases like 255, 0, 0 (red)
	case cc.G == 0 && cc.B == 0:
		cc.R = 255
	case cc.R == 0 && cc.B == 0:
		cc.G = 255
	case cc.R == 0 && cc.G == 0:
		cc.B = 255

	// For cases like 255, 255, 0 (yellow)
	case cc.R == cc.G:
		cc.R = 255
		cc.G = 255
	case cc.G == cc.B:
		cc.G = 255
		cc.B = 255
	case cc.R == cc.B:
		cc.R = 255
		cc.B = 255

	// For cases like 255, 192, 0 where a ratio is needed
	case cc.R == 0:
		if cc.G > cc.B {
			cc.B = uint8(math.Round(255 / float64(cc.G) * float64(cc.B)))
			editing = editingB
			cc.G = 255
		} else {
			cc.G = uint8(math.Round(255 / float64(cc.B) * float64(cc.G)))
			editing = editingG
			cc.B = 255
		}
	case cc.G == 0:
		if cc.B > cc.R {
			cc.R = uint8(math.Round(255 / float64(cc.B) * float64(cc.R)))
			editing = editingR
			cc.B = 255
		} else {
			cc.B = uint8(math.Round(255 / float64(cc.R) * float64(cc.B)))
			editing = editingB
			cc.R = 255
		}
	case cc.B == 0:
		if cc.R > cc.G {
			cc.G = uint8(math.Round(255 / float64(cc.R) * float64(cc.G)))
			editing = editingG
			cc.R = 255
		} else {
			cc.R = uint8(math.Round(255 / float64(cc.G) * float64(cc.R)))
			editing = editingR
			cc.G = 255
		}
	}

	var found int
	var ok bool
	incr := -1       // -1, 1, -2, 2, -3, 3 etc
	maxAttempts := 6 // TODO make global var for this, basically the step amount in the gradient values
	var find func(c rl.Color) rl.Color
	find = func(c rl.Color) rl.Color {
		found, ok = sliderColorsRev[c]
		if ok {
			// log.Println("\tfound", incr, c, found)
			return c
		}

		// log.Println("\tnot found, retrying", incr, c)
		if incr >= 0 {
			incr++
		}
		incr *= -1

		// Find the entry for the color ratio
		switch editing {
		case editingR:
			if incr >= 0 {
				if c.R+uint8(incr) <= 255 {
					c.R += uint8(incr)
				}
			} else {
				if c.R-uint8(incr*-1) >= 0 {
					c.R -= uint8(incr * -1)
				}
			}
		case editingG:
			if incr >= 0 {
				if c.G+uint8(incr) <= 255 {
					c.G += uint8(incr)
				}
			} else {
				if c.G-uint8(incr*-1) >= 0 {
					c.G -= uint8(incr * -1)
				}
			}
		case editingB:
			if incr >= 0 {
				if c.B+uint8(incr) <= 255 {
					c.B += uint8(incr)
				}
			} else {
				if c.B-uint8(incr*-1) >= 0 {
					c.B -= uint8(incr * -1)
				}
			}
		}

		if incr > maxAttempts {
			// log.Println("\tnot found, max attempts reached", incr, c)
			ok = true
			return c
		}

		return find(c)
	}
	if ok == false {
		cc = find(cc)
	}
	MoveColorSelector(found)

	// Go to the correct place in the RGB area
	var ax, ay uint8 = 255, 0
	if color.R > ay {
		ay = color.R
	}
	if color.G > ay {
		ay = color.G
	}
	if color.B > ay {
		ay = color.B
	}

	if color.R < ax {
		ax = color.R
	}
	if color.G < ax {
		ax = color.G
	}
	if color.B < ax {
		ax = color.B
	}

	var scale float32
	if ax > 0 {
		scale = float32(ay) / float32(ax)
	}
	ax = 255 - uint8(math.Ceil(float64(255/scale)))
	ay = 255 - ay
	if ax > 0 {
		ax--
	}
	if ay > 0 {
		ay--
	}
	MoveAreaSelector(int(ax), int(ay))
}

// CurrentColorSetLeftColor sets the left color and updates the UI components
// to reflect the set color
func CurrentColorSetLeftColor(color rl.Color) {
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

	SetUIHexColor(color)
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

	SetUIHexColor(color)
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
	currentColorRight = currentColorUIAddColor(CurrentFile.RightColor)
	CurrentColorSetRightColor(CurrentFile.RightColor)
	CurrentColorSetLeftColor(CurrentFile.LeftColor)

	currentColorSwap = NewButtonTexture(rl.NewRectangle(0, 0, bounds.Width/4, bounds.Width/4), "./res/icons/swap.png", false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			left := CurrentFile.LeftColor
			right := CurrentFile.RightColor
			CurrentColorSetRightColor(left)
			CurrentColorSetLeftColor(right)
			SetUIHexColor(left)
			SetUIColors(left)
		}, nil)
	currentColorBox.PushChild(currentColorSwap)

	currentColorAdd = NewButtonTexture(rl.NewRectangle(0, 0, bounds.Width/4, bounds.Width/4), "./res/icons/plus.png", false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			switch button {
			case rl.MouseLeftButton:
				PaletteUIAddColor(CurrentFile.LeftColor)
				Settings.PaletteData[0].Data = append(Settings.PaletteData[0].Data, CurrentFile.LeftColor)
				SaveSettings()
			case rl.MouseRightButton:
				PaletteUIAddColor(CurrentFile.RightColor)
				Settings.PaletteData[0].Data = append(Settings.PaletteData[0].Data, CurrentFile.RightColor)
				SaveSettings()
			}

		}, nil)
	currentColorBox.PushChild(currentColorAdd)
	currentColorBox.FlowChildren()

	return currentColorBox
}
