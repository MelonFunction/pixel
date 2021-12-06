package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

// BrushShape defines what shape the brush is
type BrushShape int

// Brush Shapes
const (
	BrushShapeSquare BrushShape = iota
	BrushShapeCircle
)

// PixelBrushTool draws a single pixel at a time and can also double as an
// eraser if eraser is true
type PixelBrushTool struct {
	lastPos                IntVec2
	name                   string
	eraser                 bool
	shouldConnectToLastPos bool
	size                   int // brush size
	shape                  BrushShape
	// Don't draw over the same pixel multiple times, prevents opacity stacking
	drawnPixels map[IntVec2]bool

	currentColor rl.Color
}

// NewPixelBrushTool returns the pixel brush tool. Requires a name and whether
// the tool is in eraser mode (helpful to prevent the current color from being
// lost)
func NewPixelBrushTool(name string, eraser bool) *PixelBrushTool {
	t := &PixelBrushTool{
		name:        name,
		eraser:      eraser,
		drawnPixels: make(map[IntVec2]bool),
		// default from File. setting manually because CurrentFile isn't set yet,
		// but it will be available on subsequent new tools
		size: 1,
	}

	if CurrentFile != nil {
		if eraser {
			t.size = CurrentFile.EraserSize
		} else {
			t.size = CurrentFile.BrushSize
		}
	}

	return t
}

func (t *PixelBrushTool) exists(e IntVec2) bool {
	_, found := t.drawnPixels[e]
	return found
}

// GetSize returns the tool size
func (t *PixelBrushTool) GetSize() int {
	if t.eraser {
		return CurrentFile.EraserSize
	}
	return CurrentFile.BrushSize
}

// SetSize sets the tool size
func (t *PixelBrushTool) SetSize(size int) {
	if size > 0 {
		t.size = size

		if t.eraser {
			CurrentFile.EraserSize = size
		} else {
			CurrentFile.BrushSize = size
		}
	}
}

// genFillShape  d is the diamater/width
func (t *PixelBrushTool) genFillShape(d int, shape BrushShape) map[IntVec2]bool {
	r := make(map[IntVec2]bool)

	switch shape {
	case BrushShapeCircle:
		// TODO
		d = d / 2
		if d < 0 {
			return nil
		}
		x, y := 0, 0
		// Bresenham algorithm
		x1, y1, err := -d, 0, 2-2*d
		for {
			r[IntVec2{x - x1, y + y1}] = true
			r[IntVec2{x - y1, y - x1}] = true
			r[IntVec2{x + x1, y - y1}] = true
			r[IntVec2{x + y1, y + x1}] = true
			d = err
			if d > x1 {
				x1++
				err += x1*2 + 1
			}
			if d <= y1 {
				y1++
				err += y1*2 + 1
			}
			if x1 >= 0 {
				break
			}
		}
	case BrushShapeSquare:
		var min, max int
		if d%2 == 0 {
			min = -d / 2
			max = d/2 - 1
		} else {
			min = -d / 2
			max = d / 2
		}
		for xx := min; xx <= max; xx++ {
			for yy := min; yy <= max; yy++ {
				// Don't draw already drawn pixels
				r[IntVec2{xx, yy}] = true
			}
		}
	default:
		panic("Shape not specified")
	}

	return r
}

// drawPixel draws the brush stroke
func (t *PixelBrushTool) drawPixel(x, y int, color rl.Color, fileDraw bool) {
	sh := t.genFillShape(t.size, BrushShapeSquare)
	for pos := range sh {
		sx, sy := x+pos.X, y+pos.Y
		if !t.exists(IntVec2{sx, sy}) {
			if fileDraw {
				CurrentFile.DrawPixel(sx, sy, color, true)
				t.drawnPixels[IntVec2{sx, sy}] = true
			} else {
				rl.DrawPixel(sx, sy, color)
			}
		}
	}
}

func (t *PixelBrushTool) isLineModifierDown() bool {
	for _, keys := range Settings.KeymapData["drawLine"] {
		allDown := true
		for _, key := range keys {
			if !rl.IsKeyDown(key) {
				allDown = false
			}
		}

		if allDown {
			return true
		}
	}
	return false
}

// MouseDown is for mouse down events
func (t *PixelBrushTool) MouseDown(x, y int, button rl.MouseButton) {
	// Assume we are in eraser mode by setting transparent as default
	t.currentColor = rl.Transparent
	if !t.eraser {
		switch button {
		case rl.MouseLeftButton:
			t.currentColor = CurrentFile.LeftColor
		case rl.MouseRightButton:
			t.currentColor = CurrentFile.RightColor
		}
	}

	if t.shouldConnectToLastPos || t.isLineModifierDown() {
		Line(t.lastPos.X, t.lastPos.Y, x, y, func(x, y int) {
			t.drawPixel(x, y, t.currentColor, true)
		})
	} else {
		t.shouldConnectToLastPos = true
		t.drawPixel(x, y, t.currentColor, true)
	}
	t.lastPos.X = x
	t.lastPos.Y = y
}

// MouseUp is for mouse up events
func (t *PixelBrushTool) MouseUp(x, y int, button rl.MouseButton) {
	t.shouldConnectToLastPos = false
	t.drawnPixels = make(map[IntVec2]bool)
	CurrentFile.GetCurrentLayer().Redraw()
}

// DrawPreview is for drawing the preview
func (t *PixelBrushTool) DrawPreview(x, y int) {
	rl.ClearBackground(rl.Transparent)

	if t.isLineModifierDown() {
		Line(t.lastPos.X, t.lastPos.Y, x, y, func(x, y int) {
			t.drawPixel(x, y, rl.NewColor(255, 255, 255, 192), false)
		})
	}

	t.drawPixel(x, y, rl.NewColor(255, 255, 255, 192), false)
}

func (t *PixelBrushTool) String() string {
	return t.name
}
