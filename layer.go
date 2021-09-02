package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

// Layer contains data for layers
type Layer struct {
	Hidden        bool
	Canvas        rl.RenderTexture2D
	Name          string
	Width, Height int

	// PixelData is the "raw" pixels map
	PixelData map[IntVec2]rl.Color
}

// Redraw redraws the layer
func (l *Layer) Redraw() {
	rl.BeginTextureMode(l.Canvas)
	rl.ClearBackground(rl.Transparent)
	for p, color := range l.PixelData {
		rl.DrawPixel(p.X, p.Y, color)
	}

	rl.EndTextureMode()
}

// Resize the layer to the specified width, height and direction
func (l *Layer) Resize(width, height int, direction ResizeDirection) {
	l.Canvas = rl.LoadRenderTexture(width, height)

	w := CurrentFile.CanvasWidth
	h := CurrentFile.CanvasHeight

	nw := width
	nh := height

	// offsets
	dx := 0
	dy := 0

	switch CurrentFile.CanvasDirectionResizePreview {
	case ResizeTL:
		dx = 0
		dy = 0
	case ResizeTC:
		dx = (w - nw) / 2
		dy = 0
	case ResizeTR:
		dx = w - nw
		dy = 0
	case ResizeCL:
		dx = 0
		dy = (h - nh) / 2
	case ResizeCC:
		dx = (w - nw) / 2
		dy = (h - nh) / 2
	case ResizeCR:
		dx = w - nw
		dy = (h - nh) / 2
	case ResizeBL:
		dx = 0
		dy = h - nh
	case ResizeBC:
		dx = (w - nw) / 2
		dy = h - nh
	case ResizeBR:
		dx = w - nw
		dy = h - nh
	}

	newPixelData := make(map[IntVec2]rl.Color)
	rl.BeginTextureMode(l.Canvas)
	rl.ClearBackground(rl.Transparent)
	for x := dx; x < w; x++ {
		for y := dy; y < h; y++ {
			if color, ok := l.PixelData[IntVec2{x, y}]; ok {
				rl.DrawPixel(x-dx, y-dy, color)
				newPixelData[IntVec2{x - dx, y - dy}] = color
			}
		}
	}
	rl.EndTextureMode()
	l.PixelData = newPixelData
	l.Width = width
	l.Height = height
}

// NewLayer returns a pointer to a new Layer
func NewLayer(width, height int, name string, fillColor rl.Color, shouldFill bool) *Layer {
	return &Layer{
		Canvas:    rl.LoadRenderTexture(width, height),
		PixelData: make(map[IntVec2]rl.Color),
		Name:      name,
		Hidden:    false,
		Width:     width,
		Height:    height,
	}
}
