package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

// SelectorTool allows for a selection to be made
type SelectorTool struct {
	firstPos, lastPos IntVec2
	firstDown         bool
	name              string
}

// NewSelectorTool returns the selector tool
func NewSelectorTool(name string) *SelectorTool {
	return &SelectorTool{
		name: name,
	}
}

func (t *SelectorTool) getClampedCoordinates(x, y int) IntVec2 {
	if x < 0 {
		x = 0
	} else if x >= CurrentFile.CanvasWidth-1 {
		x = CurrentFile.CanvasWidth - 1
	}
	if y < 0 {
		y = 0
	} else if y >= CurrentFile.CanvasHeight-1 {
		y = CurrentFile.CanvasHeight - 1
	}

	v := IntVec2{x, y}
	return v
}

// MouseDown is for mouse down events
func (t *SelectorTool) MouseDown(x, y int, button rl.MouseButton) {
	// Only get the first position after mouse has just been clicked
	if t.firstDown == false {
		t.firstDown = true
		t.firstPos = t.getClampedCoordinates(x, y)
	}
}

// MouseUp is for mouse up events
func (t *SelectorTool) MouseUp(x, y int, button rl.MouseButton) {
	t.firstDown = false
	t.lastPos = t.getClampedCoordinates(x, y)

	CurrentFile.Selection = []IntVec2{}
	for px := t.firstPos.X; px <= t.lastPos.X; px++ {
		for py := t.firstPos.Y; py <= t.lastPos.Y; py++ {
			CurrentFile.Selection = append(CurrentFile.Selection, IntVec2{px, py})
		}
	}
}

// DrawPreview is for drawing the preview
func (t *SelectorTool) DrawPreview(x, y int) {
	rl.ClearBackground(rl.Transparent)
	// Preview pixel location with a suitable color
	c := CurrentFile.GetCurrentLayer().PixelData[IntVec2{x, y}]
	avg := (c.R + c.G + c.B) / 3
	if avg > 255/2 {
		rl.DrawPixel(x, y, rl.Color{0, 0, 0, 192})
	} else {
		rl.DrawPixel(x, y, rl.Color{255, 255, 255, 192})
	}
}

func (t *SelectorTool) String() string {
	return t.name
}
