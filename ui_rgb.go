package main

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	// Hovers over the selected color/opacity
	areaSelector    *Entity
	colorSelector   *Entity
	opacitySelector *Entity

	// Components
	hexInput      *Entity
	rgbArea       *Entity
	colorSlider   *Entity
	opacitySlider *Entity

	// The color tables and their reverse lookup table
	opacityColors    = make(map[int]rl.Color)
	opacityColorsRev = make(map[rl.Color]int32)
	areaColors       = make(map[IntVec2]rl.Color)
	sliderColors     = make(map[int]rl.Color)
	sliderColorsRev  = make(map[rl.Color]int32)

	// Used by slider to set tool color when slider is moved
	lastColorLocation IntVec2

	// hexColor is the current color being displayed
	hexColor rl.Color
)

// SetUIHexColor sets the hex color label
func SetUIHexColor(color rl.Color) {
	hexColor = color

	if drawable, ok := hexInput.GetDrawable(); ok {
		if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
			drawableText.Label = ColorToHex(color)

			// set label colors
			colorString := ColorToHex(color)
			r := colorString[0:2]
			g := colorString[2:4]
			b := colorString[4:6]
			a := colorString[6:8]
			if len(drawableText.ColoredLabel) != 4 {
				drawableText.ColoredLabel = make([]TextFormatData, 5)
				drawableText.ColoredLabel[0].Color = rl.NewColor(255, 32, 32, 255)
				drawableText.ColoredLabel[1].Color = rl.NewColor(0, 255, 0, 255)
				drawableText.ColoredLabel[2].Color = rl.NewColor(54, 197, 244, 255)
				drawableText.ColoredLabel[3].Color = rl.NewColor(255, 255, 255, 255)
				drawableText.ColoredLabel[4].Color = rl.NewColor(255, 255, 255, 255)
			}
			drawableText.ColoredLabel[0].Text = r
			drawableText.ColoredLabel[1].Text = g
			drawableText.ColoredLabel[2].Text = b
			drawableText.ColoredLabel[3].Text = a
			drawableText.ColoredLabel[4].Text = ""
		}
	}
}

func makeColorArea() {
	// Setup the colorSlider texture
	if drawable, ok := colorSlider.GetDrawable(); ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			texture := renderTexture.Texture
			rl.BeginTextureMode(texture)
			w := texture.Texture.Width
			fraction := int32(w / 6)
			for px := int32(0); px < int32(texture.Texture.Width); px++ {
				// 100, 110, 010, 011, 001, 101, 100
				color := rl.NewColor(0, 0, 0, 255)

				p := (float32(px%fraction) / (float32(fraction) - 1))
				switch {
				case px >= 0 && px < fraction:
					// 100 to 110
					color.R = 255
					color.G = uint8(float32(255) * p)
				case px >= fraction && px < fraction*2:
					// 110 to 010
					color.R = uint8(float32(255) * (1 - p))
					color.G = 255
				case px >= fraction*2 && px < fraction*3:
					// 010 to 011
					color.G = 255
					color.B = uint8(float32(255) * p)
				case px >= fraction*3 && px < fraction*4:
					// 011 to 001
					color.G = uint8(float32(255) * (1 - p))
					color.B = 255
				case px >= fraction*4 && px < fraction*5:
					// 001 to 101
					color.R = uint8(float32(255) * p)
					color.B = 255
				case px >= fraction*5 && px < fraction*6:
					// 101 to 100
					color.R = 255
					color.B = uint8(float32(255) * (1 - p))
				default:
					// Fill in the remainder with red
					color.R = 255
				}

				for py := int32(0); py < int32(texture.Texture.Height); py++ {
					rl.DrawPixel(px, py, color)
					sliderColors[int(px)] = color
					sliderColorsRev[color] = px
				}
			}
			rl.EndTextureMode()
		}
	}
}

func makeSelector() *Entity {
	e := NewRenderTexture(rl.NewRectangle(-64, -64, 16, 16), nil, nil)
	if drawable, ok := e.GetDrawable(); ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			texture := renderTexture.Texture
			rl.BeginTextureMode(texture)
			rl.ClearBackground(rl.Blank)
			w := float32(texture.Texture.Width)
			h := float32(texture.Texture.Height)
			var t float32 = 3.0 // line thickness

			rl.DrawLineEx(rl.NewVector2(t, 0), rl.NewVector2(w-t, 0), t*2, rl.White) // top
			rl.DrawLineEx(rl.NewVector2(0, t), rl.NewVector2(0, h-t), t*2, rl.White) // left
			rl.DrawLineEx(rl.NewVector2(w, t), rl.NewVector2(w, h-t), t*2, rl.White) // right
			rl.DrawLineEx(rl.NewVector2(t, h), rl.NewVector2(w-t, h), t*2, rl.White) // bottom

			rl.EndTextureMode()
		}
	}
	e.Name = "selector"
	return e
}

// Generates the gradient for the color area
func makeBlendArea(origColor rl.Color) {
	if drawable, ok := rgbArea.GetDrawable(); ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			texture := renderTexture.Texture
			rl.BeginTextureMode(texture)
			w := texture.Texture.Width
			h := texture.Texture.Height

			for py := int32(0); py < int32(h); py++ {

				for px := int32(0); px < int32(w); px++ {
					color := rl.NewColor(0, 0, 0, 255)

					// Lerp from white to origColor
					ph := (float32(px) / float32(w-1))
					color.R = uint8((255*(1-ph) + float32(origColor.R)*(ph)))
					color.G = uint8((255*(1-ph) + float32(origColor.G)*(ph)))
					color.B = uint8((255*(1-ph) + float32(origColor.B)*(ph)))

					// Lerp everything to black
					pv := (float32(py) / float32(h-1))
					color.R = uint8(float32(color.R) * (1 - pv))
					color.G = uint8(float32(color.G) * (1 - pv))
					color.B = uint8(float32(color.B) * (1 - pv))

					rl.DrawPixel(px, py, color)
					areaColors[IntVec2{px, py}] = color
				}
			}
			rl.EndTextureMode()
		}
	}
}

func makeOpacitySliderArea(color rl.Color) {
	// Setup the opacitySlider texture
	if drawable, ok := opacitySlider.GetDrawable(); ok {
		renderTexture, ok := drawable.DrawableType.(*DrawableRenderTexture)
		if ok {
			texture := renderTexture.Texture
			rl.BeginTextureMode(texture)
			w := texture.Texture.Width
			fraction := int32(w)
			for px := int32(0); px < int32(texture.Texture.Width); px++ {
				drawColor := color

				p := (float32(px%fraction) / (float32(fraction) - 1))
				color.A = uint8(float32(255) * p)

				drawColor.A = 255
				drawColor.R = uint8(float32(drawColor.R) * p)
				drawColor.G = uint8(float32(drawColor.G) * p)
				drawColor.B = uint8(float32(drawColor.B) * p)

				for py := int32(0); py < int32(texture.Texture.Height); py++ {
					rl.DrawPixel(px, py, drawColor)
					opacityColors[int(px)] = color
					opacityColorsRev[color] = px
				}
			}
			rl.EndTextureMode()
		}
	}
}

// MoveAreaSelector moves the area selector
func MoveAreaSelector(mx, my float32) {
	if moveable, ok := rgbArea.GetMoveable(); ok {
		if mx < 0 {
			mx = 0
		}
		if my < 0 {
			my = 0
		}
		if mx > 1 {
			mx = 1
		}
		if my > 1 {
			my = 1
		}

		// Move the areaSelector
		if sm, ok := areaSelector.GetMoveable(); ok {
			sm.Bounds.X = moveable.Bounds.X + moveable.Bounds.Width*mx - sm.Bounds.Width/2
			sm.Bounds.Y = moveable.Bounds.Y + moveable.Bounds.Height*my - sm.Bounds.Height/2
		}
	}
}

// MoveColorSelector moves the color selector
func MoveColorSelector(mx int32) {
	if moveable, ok := colorSlider.GetMoveable(); ok {
		my := int32(moveable.Bounds.Height) / 2

		if mx < 0 {
			mx = 0
		}
		if mx > int32(moveable.Bounds.Width)-1 {
			mx = int32(moveable.Bounds.Width) - 1
		}

		// Move the colorSelector
		if sm, ok := colorSelector.GetMoveable(); ok {
			sm.Bounds.X = moveable.Bounds.X + float32(mx) - sm.Bounds.Width/2
			sm.Bounds.Y = moveable.Bounds.Y + float32(my) - sm.Bounds.Height/2
		}
	}
}

// MoveOpacitySelector moves the color selector
func MoveOpacitySelector(percent float32) {
	if moveable, ok := opacitySlider.GetMoveable(); ok {
		my := int32(moveable.Bounds.Height) / 2

		if percent < 0 {
			percent = 0
		}
		if percent > 1 {
			percent = 1
		}

		pos := moveable.Bounds.Width * percent

		// Move the opacitySelector
		if sm, ok := opacitySelector.GetMoveable(); ok {
			sm.Bounds.X = moveable.Bounds.X + pos - sm.Bounds.Width/2
			sm.Bounds.Y = moveable.Bounds.Y + float32(my) - sm.Bounds.Height/2
		}
	}
}

// TODO keep opacity on color change
// TODO move selectors on start

// NewRGBUI creates the UI representation of the color picker
func NewRGBUI(bounds rl.Rectangle) *Entity {
	// The main color gradient area, fading from white to the current color
	// horizontally, then vertically down to black
	areaBounds := bounds
	areaBounds.Height = areaBounds.Width
	rgbArea = NewRenderTexture(areaBounds,
		func(entity *Entity, button MouseButton) {
			// button up
		},
		func(entity *Entity, button MouseButton, isHeld bool) {
			// button down
			PaletteUIHideCurrentColorIndicator()
			if moveable, ok := rgbArea.GetMoveable(); ok {
				mx := rl.GetMouseX()
				my := rl.GetMouseY()
				mx -= int32(moveable.Bounds.X)
				my -= int32(moveable.Bounds.Y)

				if mx < 0 {
					mx = 0
				}
				if my < 0 {
					my = 0
				}
				if mx > int32(moveable.Bounds.Width)-1 {
					mx = int32(moveable.Bounds.Width) - 1
				}
				if my > int32(moveable.Bounds.Height)-1 {
					my = int32(moveable.Bounds.Height) - 1
				}

				// Move the areaSelector
				if sm, ok := areaSelector.GetMoveable(); ok {
					sm.Bounds.X = moveable.Bounds.X + float32(mx) - sm.Bounds.Width/2
					sm.Bounds.Y = moveable.Bounds.Y + float32(my) - sm.Bounds.Height/2
				}

				loc := IntVec2{mx, my}
				color, ok := areaColors[loc]
				if ok {
					// Set the current color in the file
					lastColorLocation = loc

					// MoveOpacitySelector(float32(color.A) / 255)
					switch button {
					case rl.MouseLeftButton:
						color.A = LeftColor.A
						CurrentColorSetLeftColor(color)
					case rl.MouseRightButton:
						color.A = RightColor.A
						CurrentColorSetRightColor(color)
					}

					makeOpacitySliderArea(color)
				}
			}
		},
	)

	// The slider for colors
	sliderBounds := bounds
	sliderBounds.Height = UIButtonHeight / 2
	colorSlider = NewRenderTexture(sliderBounds,
		func(entity *Entity, button MouseButton) {
			// button up
		},
		func(entity *Entity, button MouseButton, isHeld bool) {
			// button down
			PaletteUIHideCurrentColorIndicator()
			if moveable, ok := colorSlider.GetMoveable(); ok {
				mx := rl.GetMouseX()
				mx -= int32(moveable.Bounds.X)
				my := int32(moveable.Bounds.Height) / 2

				if mx < 0 {
					mx = 0
				}
				if mx > int32(moveable.Bounds.Width)-1 {
					mx = int32(moveable.Bounds.Width) - 1
				}

				// Move the colorSelector
				if sm, ok := colorSelector.GetMoveable(); ok {
					sm.Bounds.X = moveable.Bounds.X + float32(mx) - sm.Bounds.Width/2
					sm.Bounds.Y = moveable.Bounds.Y + float32(my) - sm.Bounds.Height/2
				}

				color, ok := sliderColors[int(mx)]
				if ok {
					if c, ok := areaColors[lastColorLocation]; ok {
						// Update current color with the position-adjusted color
						switch button {
						case rl.MouseLeftButton:
							c.A = LeftColor.A
							CurrentColorSetLeftColor(c)
						case rl.MouseRightButton:
							c.A = RightColor.A
							CurrentColorSetRightColor(c)
						}
					}

					// Update color areas with the "raw" color from the slider
					makeBlendArea(color)
					makeOpacitySliderArea(color)
				}
			}
		},
	)

	// The slider for opacity
	opacitySlider = NewRenderTexture(sliderBounds,
		func(entity *Entity, button MouseButton) {
			// button up
		},
		func(entity *Entity, button MouseButton, isHeld bool) {
			// button down
			if moveable, ok := opacitySlider.GetMoveable(); ok {
				mx := rl.GetMouseX()
				mx -= int32(moveable.Bounds.X)
				my := int32(moveable.Bounds.Height) / 2

				if mx < 0 {
					mx = 0
				}
				if mx > int32(moveable.Bounds.Width)-1 {
					mx = int32(moveable.Bounds.Width) - 1
				}

				// Move the opacitySelector
				if sm, ok := opacitySelector.GetMoveable(); ok {
					sm.Bounds.X = moveable.Bounds.X + float32(mx) - sm.Bounds.Width/2
					sm.Bounds.Y = moveable.Bounds.Y + float32(my) - sm.Bounds.Height/2
				}

				if color, ok := opacityColors[int(mx)]; ok {
					switch button {
					case rl.MouseLeftButton:
						LeftColor.A = color.A
						CurrentColorSetLeftColor(LeftColor)
					case rl.MouseRightButton:
						RightColor.A = color.A
						CurrentColorSetRightColor(RightColor)
					}
				}
			}
		},
	)

	// Generate the initial textures
	makeBlendArea(rl.NewColor(255, 0, 0, 255))
	makeOpacitySliderArea(rl.Red)
	makeColorArea()

	hexColor = LeftColor
	hexInput = NewInput(sliderBounds, "#00000000", TextAlignCenter, false, func(entity *Entity, button MouseButton) {}, nil,
		func(entity *Entity, key Key) {
			if drawable, ok := entity.GetDrawable(); ok {
				if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {

					if key == rl.KeyEnter {
						RemoveCapturedInput()
					} else if key == rl.KeyBackspace && len(drawableText.Label) > 0 {
						drawableText.Label = drawableText.Label[:len(drawableText.Label)-1]
					} else if len(drawableText.Label) < 8 {
						switch {
						case key >= 48 && key <= 57: // 0 to 9
							fallthrough
						case key >= 97 && key <= 102: // a to f
							fallthrough
						case key >= rl.KeyA && key <= rl.KeyF:
							drawableText.Label += strings.ToLower(string(rune(key)))
						}
					}

					// Set the color from the hex
					if color, err := HexToColor(drawableText.Label); err == nil {
						hexColor = color
						SetUIColors(hexColor)
						makeBlendArea(color)
						makeOpacitySliderArea(color)
					}

					// clear all label positions (if this is in next loop, chars get stranded on backspace)
					for i := 0; i <= 3; i++ {
						drawableText.ColoredLabel[i].Text = ""
					}
					for i := 0; i < len(drawableText.Label); i += 2 {
						text := ""
						if i+2 <= len(drawableText.Label) {
							text = drawableText.Label[i : i+2]
							// log.Printf("%d - %d:%d %s\n", i/2, i, i+2, text)
						} else if i+1 <= len(drawableText.Label) {
							text = drawableText.Label[i : i+1]
							// log.Printf("%d - %d:%d %s\n", i/2, i, i+1, text)
						}

						drawableText.ColoredLabel[i/2].Text = text
					}
				}
			}
		})
	if interactable, ok := hexInput.GetInteractable(); ok {
		interactable.OnBlur = func(entity *Entity) {
			// SetUIColors(hexColor)
			CurrentColorSetLeftColor(hexColor)
			if drawable, ok := hexInput.GetDrawable(); ok {
				if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
					// drawableText.ColoredLabel = nil
					drawableText.ColoredLabel[4].Text = ""
				}
			}
		}
		interactable.OnFocus = func(entity *Entity) {
			// just render normal label when editing, logic is too annoying

			if drawable, ok := hexInput.GetDrawable(); ok {
				if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
					// drawableText.ColoredLabel = nil
					drawableText.ColoredLabel[4].Text = "_"
				}
			}

		}
	}

	container := NewBox(bounds, []*Entity{
		rgbArea,
		colorSlider,
		opacitySlider,
		hexInput,
	}, FlowDirectionVertical)
	container.FlowChildren()

	// Make the selector which floats around on top of the color gradient area
	// Also move it off screen for now TODO starting position depending on starting color
	areaSelector = makeSelector()
	colorSelector = makeSelector()
	opacitySelector = makeSelector()

	return container
}
