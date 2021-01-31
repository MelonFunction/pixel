package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

func NewRGBUI(bounds rl.Rectangle) *Entity {
	rgbBounds := bounds
	rgbBounds.Height = rgbBounds.Width
	var rgb *Entity
	rgb = NewRenderTexture(rgbBounds,
		func(entity *Entity, button rl.MouseButton) {
			// button up
		},
		func(entity *Entity, button rl.MouseButton) {
			// button down
		})

	makeBlendArea := func(origColor rl.Color) {
		if res, err := scene.QueryID(rgb.ID); err == nil {
			drawable := res.Components[rgb.Scene.ComponentsMap["drawable"]].(*Drawable)
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
					}
				}
				rl.EndTextureMode()
			}
		}
	}

	makeBlendArea(rl.NewColor(255, 0, 0, 255))

	sliderBounds := bounds
	sliderBounds.Height = bounds.Height - rgbBounds.Height
	var slider *Entity
	var sliderColors = make(map[int]rl.Color)
	slider = NewRenderTexture(sliderBounds,
		func(entity *Entity, button rl.MouseButton) {
			// button up
		},
		func(entity *Entity, button rl.MouseButton) {
			// button down
			if res, err := scene.QueryID(slider.ID); err == nil {
				moveable := res.Components[slider.Scene.ComponentsMap["moveable"]].(*Moveable)
				mx := rl.GetMouseX()
				mx -= int(moveable.Bounds.X)

				color, ok := sliderColors[mx]
				if ok {
					makeBlendArea(color)
				}
			}

		})

	if res, err := scene.QueryID(slider.ID); err == nil {
		drawable := res.Components[slider.Scene.ComponentsMap["drawable"]].(*Drawable)
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
				}

				for py := 0; py < int(texture.Texture.Height); py++ {
					rl.DrawPixel(px, py, color)
					sliderColors[px] = color
				}
			}
			rl.EndTextureMode()
		}
	}

	_ = rgb
	_ = slider
	container := NewBox(bounds, []*Entity{
		rgb,
		slider,
	}, FlowDirectionVertical)
	return container
}
