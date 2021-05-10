package main

import (
	"fmt"
	"log"
	"strconv"

	rl "github.com/lachee/raylib-goplus/raylib"
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
	opacityColorsRev = make(map[rl.Color]int)
	areaColors       = make(map[IntVec2]rl.Color)
	sliderColors     = make(map[int]rl.Color)
	sliderColorsRev  = make(map[rl.Color]int)

	// Used by slider to set tool color when slider is moved
	lastColorLocation IntVec2

	// hexColor is the current color being displayed
	hexColor rl.Color
)

func SetUIHexColor(color rl.Color) {
	hexColor = color

	if drawable, ok := hexInput.GetDrawable(); ok {
		if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
			drawableText.Label = fmt.Sprintf("%02x%02x%02x%02x", color.R, color.G, color.B, color.A)
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
			fraction := int(w / 6)
			for px := 0; px < int(texture.Texture.Width); px++ {
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

				for py := 0; py < int(texture.Texture.Height); py++ {
					rl.DrawPixel(px, py, color)
					sliderColors[px] = color
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
			rl.ClearBackground(rl.Transparent)
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

			for py := 0; py < int(h); py++ {

				for px := 0; px < int(w); px++ {
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
			fraction := int(w)
			for px := 0; px < int(texture.Texture.Width); px++ {
				drawColor := color

				p := (float32(px%fraction) / (float32(fraction) - 1))
				color.A = uint8(float32(255) * p)

				drawColor.A = 255
				drawColor.R = uint8(float32(drawColor.R) * p)
				drawColor.G = uint8(float32(drawColor.G) * p)
				drawColor.B = uint8(float32(drawColor.B) * p)

				for py := 0; py < int(texture.Texture.Height); py++ {
					rl.DrawPixel(px, py, drawColor)
					opacityColors[px] = color
					opacityColorsRev[color] = px
				}
			}
			rl.EndTextureMode()
		}
	}
}

func MoveAreaSelector(mx, my int) {
	if moveable, ok := rgbArea.GetMoveable(); ok {
		if mx < 0 {
			mx = 0
		}
		if my < 0 {
			my = 0
		}
		if mx > int(moveable.Bounds.Width)-1 {
			mx = int(moveable.Bounds.Width) - 1
		}
		if my > int(moveable.Bounds.Height)-1 {
			my = int(moveable.Bounds.Height) - 1
		}

		// Move the areaSelector
		if sm, ok := areaSelector.GetMoveable(); ok {
			sm.Bounds.X = moveable.Bounds.X + float32(mx) - sm.Bounds.Width/2
			sm.Bounds.Y = moveable.Bounds.Y + float32(my) - sm.Bounds.Height/2
		}

		loc := IntVec2{mx, my}
		color, ok := areaColors[loc]
		if ok {
			lastColorLocation = loc
			makeOpacitySliderArea(color)
		}
	}
}

func MoveColorSelector(mx int) {
	if moveable, ok := colorSlider.GetMoveable(); ok {
		my := int(moveable.Bounds.Height) / 2

		if mx < 0 {
			mx = 0
		}
		if mx > int(moveable.Bounds.Width)-1 {
			mx = int(moveable.Bounds.Width) - 1
		}

		// Move the colorSelector
		if sm, ok := colorSelector.GetMoveable(); ok {
			sm.Bounds.X = moveable.Bounds.X + float32(mx) - sm.Bounds.Width/2
			sm.Bounds.Y = moveable.Bounds.Y + float32(my) - sm.Bounds.Height/2
		}

		color, ok := sliderColors[mx]
		if ok {
			makeBlendArea(color)
			makeOpacitySliderArea(color)
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
		func(entity *Entity, button rl.MouseButton) {
			// button up
		},
		func(entity *Entity, button rl.MouseButton, isHeld bool) {
			// button down
			if moveable, ok := rgbArea.GetMoveable(); ok {
				mx := rl.GetMouseX()
				my := rl.GetMouseY()
				mx -= int(moveable.Bounds.X)
				my -= int(moveable.Bounds.Y)

				if mx < 0 {
					mx = 0
				}
				if my < 0 {
					my = 0
				}
				if mx > int(moveable.Bounds.Width)-1 {
					mx = int(moveable.Bounds.Width) - 1
				}
				if my > int(moveable.Bounds.Height)-1 {
					my = int(moveable.Bounds.Height) - 1
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

					switch button {
					case rl.MouseLeftButton:
						CurrentColorSetLeftColor(color)
					case rl.MouseRightButton:
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
		func(entity *Entity, button rl.MouseButton) {
			// button up
		},
		func(entity *Entity, button rl.MouseButton, isHeld bool) {
			// button down
			if moveable, ok := colorSlider.GetMoveable(); ok {
				mx := rl.GetMouseX()
				mx -= int(moveable.Bounds.X)
				my := int(moveable.Bounds.Height) / 2

				if mx < 0 {
					mx = 0
				}
				if mx > int(moveable.Bounds.Width)-1 {
					mx = int(moveable.Bounds.Width) - 1
				}

				// Move the colorSelector
				if sm, ok := colorSelector.GetMoveable(); ok {
					sm.Bounds.X = moveable.Bounds.X + float32(mx) - sm.Bounds.Width/2
					sm.Bounds.Y = moveable.Bounds.Y + float32(my) - sm.Bounds.Height/2
				}

				color, ok := sliderColors[mx]
				if ok {
					makeBlendArea(color)
					makeOpacitySliderArea(color)

					// Update the current color with the last color location
					color, ok := areaColors[lastColorLocation]
					if ok {
						// Set the current color in the file
						switch button {
						case rl.MouseLeftButton:
							CurrentColorSetLeftColor(color)
						case rl.MouseRightButton:
							CurrentColorSetRightColor(color)
						}
					}
				}
			}
		},
	)

	// The slider for opacity
	opacitySlider = NewRenderTexture(sliderBounds,
		func(entity *Entity, button rl.MouseButton) {
			// button up
		},
		func(entity *Entity, button rl.MouseButton, isHeld bool) {
			// button down
			if moveable, ok := opacitySlider.GetMoveable(); ok {
				mx := rl.GetMouseX()
				mx -= int(moveable.Bounds.X)
				my := int(moveable.Bounds.Height) / 2

				if mx < 0 {
					mx = 0
				}
				if mx > int(moveable.Bounds.Width)-1 {
					mx = int(moveable.Bounds.Width) - 1
				}

				// Move the opacitySelector
				if sm, ok := opacitySelector.GetMoveable(); ok {
					sm.Bounds.X = moveable.Bounds.X + float32(mx) - sm.Bounds.Width/2
					sm.Bounds.Y = moveable.Bounds.Y + float32(my) - sm.Bounds.Height/2
				}

				if color, ok := opacityColors[mx]; ok {
					switch button {
					case rl.MouseLeftButton:
						CurrentColorSetLeftColor(color)
					case rl.MouseRightButton:
						CurrentColorSetRightColor(color)
					}
				}
			}
		},
	)

	// Generate the initial textures
	makeBlendArea(rl.NewColor(255, 0, 0, 255))
	makeOpacitySliderArea(rl.Red)
	makeColorArea()

	hexColor = CurrentFile.LeftColor
	hexInput = NewInput(sliderBounds, "#00000000", false, func(entity *Entity, button rl.MouseButton) {}, nil,
		func(entity *Entity, key rl.Key) {
			if drawable, ok := entity.GetDrawable(); ok {
				// TODO pasting
				// TODO set color from hex code

				if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
					if key == rl.KeyBackspace && len(drawableText.Label) > 0 {
						drawableText.Label = drawableText.Label[:len(drawableText.Label)-1]
					} else if len(drawableText.Label) < 8 {
						switch {
						// 0 to 9
						case key >= 48 && key <= 57:
							fallthrough
						// a to f
						case key >= 97 && key <= 102:
							fallthrough
						case key >= rl.KeyA && key <= rl.KeyF:
							drawableText.Label += string(rune(key))
						}
					}

					// Set the color from the hex
					var r, g, b, a int64 = 0, 0, 0, 255
					var err error
					switch len(drawableText.Label) {
					case 8:
						if a, err = strconv.ParseInt(drawableText.Label[6:8], 16, 32); err != nil {
							log.Println(err)
						}
						fallthrough
					case 6:
						if r, err = strconv.ParseInt(drawableText.Label[0:2], 16, 32); err != nil {
							log.Println(err)
						}
						if g, err = strconv.ParseInt(drawableText.Label[2:4], 16, 32); err != nil {
							log.Println(err)
						}
						if b, err = strconv.ParseInt(drawableText.Label[4:6], 16, 32); err != nil {
							log.Println(err)
						}

						hexColor = rl.Color{uint8(r), uint8(g), uint8(b), uint8(a)}
					}
				}
			}
		})
	if interactable, ok := hexInput.GetInteractable(); ok {
		interactable.OnBlur = func(entity *Entity) {
			SetUIColors(hexColor)
			CurrentColorSetLeftColor(hexColor)
		}
		interactable.OnFocus = func(entity *Entity) {
			SetUIHexColor(hexColor)
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
