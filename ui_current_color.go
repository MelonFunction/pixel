package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	// The container
	currentColorBox *Entity

	currentColorLeft  *Entity
	currentColorRight *Entity
	currentColorSwap  *Entity
	currentColorAdd   *Entity

	currentColorPlusTexture     rl.Texture2D
	currentColorNegativeTexture rl.Texture2D
)

type colorEditing int32

const (
	editingR colorEditing = iota
	editingG
	editingB
)

// SetUIColors moves the pointers in the color areas
// This should only be used for cases like selecting a color from the palette, switching colors etc as it isn't
// compatible with selecting directly from the color area
func SetUIColors(color rl.Color) {
	cc := color
	cc.A = 255

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

	var foundColor int32
	_ = foundColor
	var ok bool
	incr := -1       // -1, 1, -2, 2, -3, 3 etc
	maxAttempts := 6 // TODO make global var for this, basically the step amount in the gradient values
	var find func(c rl.Color) rl.Color
	find = func(c rl.Color) rl.Color {
		foundColor, ok = sliderColorsRev[c]
		if ok {
			// log.Println("\tfoundColor", incr, c, foundColor)
			return c
		}

		// log.Println("\tnot foundColor, retrying", incr, c)
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
			// log.Println("\tnot foundColor, max attempts reached", incr, c)
			ok = true
			return c
		}

		return find(c)
	}
	if ok == false {
		cc = find(cc)
	}

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

	MoveColorSelector(foundColor)
	MoveOpacitySelector(float32(color.A) / 255)
	MoveAreaSelector(float32(ax)/255, float32(ay)/255)
}

// CurrentColorSetLeftColor sets the left color and updates the UI components
// to reflect the set color
func CurrentColorSetLeftColor(color rl.Color) {
	if drawable, ok := currentColorLeft.GetDrawable(); ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			LeftColor = color

			texture := renderTexture.Texture
			rl.BeginTextureMode(texture)
			rl.ClearBackground(color)
			if int32(color.R)+int32(color.G)+int32(color.B) < 128 || color.A < 128 {
				rl.DrawRectangleLinesEx(rl.NewRectangle(0, 0, float32(renderTexture.Texture.Texture.Width), float32(renderTexture.Texture.Texture.Height)), 2, rl.Gray)
			}
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
			RightColor = color

			texture := renderTexture.Texture
			rl.BeginTextureMode(texture)
			rl.ClearBackground(color)
			if int32(color.R)+int32(color.G)+int32(color.B) < 128 || color.A < 128 {
				rl.DrawRectangleLinesEx(rl.NewRectangle(0, 0, float32(renderTexture.Texture.Texture.Width), float32(renderTexture.Texture.Texture.Height)), 2, rl.Gray)
			}
			rl.EndTextureMode()
		}
	}

	SetUIColors(color)
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

// CurrentColorToggleAddRemoveGraphic changes the texture of the button which adds/removes the currently selected color
// to the current palette
func CurrentColorToggleAddRemoveGraphic() {
	drawable, ok := currentColorAdd.GetDrawable()
	if ok {
		dt, ok := drawable.DrawableType.(*DrawableTexture)
		if ok {
			if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
				dt.Texture = currentColorNegativeTexture
			} else {
				dt.Texture = currentColorPlusTexture
			}
		}
	}
}

// NewCurrentColorUI creates a new Current Color UI component which displays
// the currently selected colors as well as buttons to swap them and add the
// color to the palette depending on which mouse button was clicked
func NewCurrentColorUI(bounds rl.Rectangle) *Entity {
	currentColorPlusTexture = rl.LoadTexture("./res/icons/plus.png")
	currentColorNegativeTexture = rl.LoadTexture("./res/icons/negative.png")

	currentColorBox = NewBox(bounds, []*Entity{}, FlowDirectionHorizontal)

	currentColorLeft = currentColorUIAddColor(LeftColor)
	currentColorRight = currentColorUIAddColor(RightColor)
	CurrentColorSetRightColor(RightColor)
	CurrentColorSetLeftColor(LeftColor)

	currentColorSwap = NewButtonTexture(rl.NewRectangle(0, 0, bounds.Width/4, bounds.Width/4), GetFile("./res/icons/swap.png"), false,
		func(entity *Entity, button MouseButton) {
			// button up
			left := LeftColor
			right := RightColor
			CurrentColorSetRightColor(left)
			CurrentColorSetLeftColor(right)
			SetUIHexColor(LeftColor)
			// SetUIColors(LeftColor)
		}, nil)
	currentColorBox.PushChild(currentColorSwap)

	currentColorAdd = NewButtonTexture(rl.NewRectangle(0, 0, bounds.Width/4, bounds.Width/4), GetFile("./res/icons/plus.png"), false,
		func(entity *Entity, button MouseButton) {
			// button up
			if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
				// remove color

				if colors, err := PaletteUIPaletteEntity.GetChildren(); err == nil {
					for index, color := range colors {
						if color == PaletteUICurrentColorEntity {
							PaletteUIRemoveColor(PaletteUICurrentColorEntity)
							PaletteUIPreviousColor()
							Settings.PaletteData[CurrentFile.CurrentPalette].data = append(
								Settings.PaletteData[CurrentFile.CurrentPalette].data[:index],
								Settings.PaletteData[CurrentFile.CurrentPalette].data[index+1:]...,
							)
							SaveSettings()
							return
						}
					}

				}
			} else {
				// add color
				switch button {
				case rl.MouseLeftButton:
					PaletteUIAddColor(LeftColor, int32(len(Settings.PaletteData[CurrentFile.CurrentPalette].Strings)))
					Settings.PaletteData[CurrentFile.CurrentPalette].data = append(Settings.PaletteData[CurrentFile.CurrentPalette].data, LeftColor)
					SaveSettings()
				case rl.MouseRightButton:
					PaletteUIAddColor(RightColor, int32(len(Settings.PaletteData[CurrentFile.CurrentPalette].Strings)))
					Settings.PaletteData[CurrentFile.CurrentPalette].data = append(Settings.PaletteData[CurrentFile.CurrentPalette].data, RightColor)
					SaveSettings()
				}
			}

		}, nil)
	currentColorBox.PushChild(currentColorAdd)
	currentColorBox.FlowChildren()

	return currentColorBox
}
