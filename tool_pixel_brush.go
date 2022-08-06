package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// BrushShape defines what shape the brush is
type BrushShape int32

// Brush Shapes
const (
	BrushShapeSquare BrushShape = iota
	BrushShapeCircle
)

// Vars
const (
	maxBrushSize = 8 // inclusive
)

// Couldn't find a good way to generate circles, this works though 🥇
var circlesRaw = [][][]int8{
	1: {{1}},
	2: {{1, 1}, {1, 1}},
	3: {{0, 1, 0}, {1, 1, 1}, {0, 1, 0}},
	4: {{0, 1, 1, 0}, {1, 1, 1, 1}, {1, 1, 1, 1}, {0, 1, 1, 0}},
	5: {{0, 1, 1, 1, 0}, {1, 1, 1, 1, 1}, {1, 1, 1, 1, 1}, {1, 1, 1, 1, 1}, {0, 1, 1, 1, 0}},
	6: {
		{0, 1, 1, 1, 1, 0},
		{1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1},
		{0, 1, 1, 1, 1, 0}},
	7: {
		{0, 0, 1, 1, 1, 0, 0},
		{0, 1, 1, 1, 1, 1, 0},
		{1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1},
		{0, 1, 1, 1, 1, 1, 0},
		{0, 0, 1, 1, 1, 0, 0}},
	8: {
		{0, 0, 1, 1, 1, 1, 0, 0},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{0, 0, 1, 1, 1, 1, 0, 0}},
}

// PixelBrushTool draws a single pixel at a time and can also double as an
// eraser if eraser is true
type PixelBrushTool struct {
	lastPos                IntVec2
	name                   string
	eraser                 bool
	shouldConnectToLastPos bool
	size                   int32 // brush size
	shape                  BrushShape
	// Don't draw over the same pixel multiple times, prevents opacity stacking
	drawnPixels map[IntVec2]bool

	currentColor rl.Color
	circles      []map[IntVec2]bool
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
		size:    1,
		circles: make([]map[IntVec2]bool, maxBrushSize+1),
	}

	for d, c := range circlesRaw {
		circle := make(map[IntVec2]bool)

		var min int
		if d%2 == 0 {
			min = -d/2 + 1
		} else {
			min = -d / 2
		}
		for xx := 0; xx < d; xx++ {
			for yy := 0; yy < d; yy++ {
				// if d%2 == 0 {
				// 	circle[IntVec2{int32(xx) + 1, int32(yy) + 1}] = true
				// } else {
				// }
				loc := IntVec2{int32(xx + min), int32(yy + min)}
				if c[yy][xx] == 1 {
					circle[loc] = true
				}
			}
		}
		t.circles[d] = circle
	}

	if eraser {
		t.size = GlobalEraserSize
		t.shape = GlobalErasorShape
	} else {
		t.size = GlobalBrushSize
		t.shape = GlobalBrushShape
	}

	return t
}

func (t *PixelBrushTool) exists(e IntVec2) bool {
	_, found := t.drawnPixels[e]
	return found
}

// GetSize returns the tool size
func (t *PixelBrushTool) GetSize() int32 {
	if t.eraser {
		return GlobalEraserSize
	}
	return GlobalBrushSize
}

// SetSize sets the tool size
func (t *PixelBrushTool) SetSize(size int32) {
	if size > 0 && size <= maxBrushSize {
		t.size = size

		if t.eraser {
			GlobalEraserSize = size
		} else {
			GlobalBrushSize = size
		}
	}
}

// GetShape returns the tool shape
func (t *PixelBrushTool) GetShape() BrushShape {
	if t.eraser {
		return GlobalErasorShape
	}
	return GlobalBrushShape
}

// SetShape sets the tool shape
func (t *PixelBrushTool) SetShape(shape BrushShape) {
	t.shape = shape
	if t.eraser {
		GlobalErasorShape = shape
	} else {
		GlobalBrushShape = shape
	}
}

// genFillShape d is the diamater/width
func (t *PixelBrushTool) genFillShape(d int32, shape BrushShape) map[IntVec2]bool {
	r := make(map[IntVec2]bool)

	switch shape {
	case BrushShapeCircle:
		r = t.circles[d]
		// r[IntVec2{0, 0}] = true
	case BrushShapeSquare:
		var min, max int32
		if d%2 == 0 {
			min = -d / 2
			max = d/2 - 1
		} else {
			min = -d / 2
			max = d / 2
		}
		for xx := min; xx <= max; xx++ {
			for yy := min; yy <= max; yy++ {
				if d%2 == 0 {
					r[IntVec2{xx + 1, yy + 1}] = true
				} else {
					r[IntVec2{xx, yy}] = true
				}
			}
		}
	default:
		panic("Shape not specified")
	}

	return r
}

// drawPixel draws the brush stroke
func (t *PixelBrushTool) drawPixel(x, y int32, color rl.Color, fileDraw bool) {
	sh := t.genFillShape(t.size, t.shape)
	for pos := range sh {
		sx, sy := x+pos.X, y+pos.Y
		if !t.exists(IntVec2{sx, sy}) {
			if fileDraw {
				CurrentFile.DrawPixel(sx, sy, color, CurrentFile.GetCurrentLayer())
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
			if !rl.IsKeyDown(int32(key)) {
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
func (t *PixelBrushTool) MouseDown(x, y int32, button MouseButton) {
	// Assume we are in eraser mode by setting transparent as default
	t.currentColor = rl.Blank
	if !t.eraser {
		switch button {
		case rl.MouseLeftButton:
			t.currentColor = LeftColor
		case rl.MouseRightButton:
			t.currentColor = RightColor
		}
	}

	if t.shouldConnectToLastPos || t.isLineModifierDown() {
		Line(t.lastPos.X, t.lastPos.Y, x, y, func(x, y int32) {
			// prevent drawing over the first pixel and stacking them, with color.A<255, opacity stacks 😠
			if !(x == t.lastPos.X && y == t.lastPos.Y) {
				t.drawPixel(x, y, t.currentColor, true)
			}
		})
	} else {
		t.shouldConnectToLastPos = true
		t.drawPixel(x, y, t.currentColor, true)
	}
	t.lastPos.X = x
	t.lastPos.Y = y
}

// MouseUp is for mouse up events
func (t *PixelBrushTool) MouseUp(x, y int32, button MouseButton) {
	t.shouldConnectToLastPos = false
	t.drawnPixels = make(map[IntVec2]bool)
	CurrentFile.GetCurrentLayer().Redraw()
}

// DrawPreview is for drawing the preview
func (t *PixelBrushTool) DrawPreview(x, y int32) {
	rl.ClearBackground(rl.Blank)

	if t.isLineModifierDown() {
		Line(t.lastPos.X, t.lastPos.Y, x, y, func(x, y int32) {
			t.drawPixel(x, y, rl.NewColor(255, 255, 255, 192), false)
		})
	}

	t.drawPixel(x, y, rl.NewColor(255, 255, 255, 192), false)
}

// DrawUI is for drawing the UI
func (t *PixelBrushTool) DrawUI(camera rl.Camera2D) {

}

func (t *PixelBrushTool) String() string {
	return t.name
}
