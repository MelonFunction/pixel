package main

import rl "github.com/lachee/raylib-goplus/raylib"

// Layer has a Canvas and hasInitialFill keeps track of if it's been filled on
// creation
type Layer struct {
	Hidden         bool
	Canvas         rl.RenderTexture2D
	hasInitialFill bool
	Name           string

	// PixelData is the "raw" pixels map
	PixelData map[IntVec2]rl.Color
}

func NewLayer(width, height int, name string, shouldFill bool) *Layer {
	return &Layer{
		Canvas:         rl.LoadRenderTexture(width, height),
		hasInitialFill: !shouldFill,
		PixelData:      make(map[IntVec2]rl.Color),
		Name:           name,
		Hidden:         false,
	}
}
