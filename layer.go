package main

import rl "github.com/lachee/raylib-goplus/raylib"

// Layer has a Canvas and initialFill keeps track of if it's been filled on
// creation
type Layer struct {
	Canvas      rl.RenderTexture2D
	initialFill bool

	// PixelData is the "raw" pixels map
	PixelData map[IntVec2]rl.Color
}

func NewLayer(width, height int, shouldFill bool) *Layer {
	return &Layer{
		Canvas:      rl.LoadRenderTexture(width, height),
		initialFill: shouldFill,
		PixelData:   make(map[IntVec2]rl.Color),
	}
}
