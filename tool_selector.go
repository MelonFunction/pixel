package main

import (
	"time"

	rl "github.com/lachee/raylib-goplus/raylib"
)

// SelectorTool allows for a selection to be made
type SelectorTool struct {
	firstPos, lastPos IntVec2
	firstDown         bool
	// Cancels the selection if a click happens without drag
	firstDownTime time.Time
	name          string
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
	CurrentFile.DoingSelection = true
	if t.firstDown == false {
		t.firstDown = true
		t.firstDownTime = time.Now()
		t.firstPos = t.getClampedCoordinates(x, y)
	} else {
		CurrentFile.Selection = []IntVec2{}
		t.lastPos = t.getClampedCoordinates(x, y)

		// Cancel selection if a click without a drag happens
		if t.firstPos.X == t.lastPos.X && t.firstPos.Y == t.lastPos.Y {
			if time.Now().Sub(t.firstDownTime) < time.Millisecond*100 {
				CurrentFile.DoingSelection = false
				return
			}
		}

		firstPosClone := t.firstPos

		if t.lastPos.X < firstPosClone.X {
			t.lastPos.X, firstPosClone.X = firstPosClone.X, t.lastPos.X
		}
		if t.lastPos.Y < firstPosClone.Y {
			t.lastPos.Y, firstPosClone.Y = firstPosClone.Y, t.lastPos.Y
		}

		for py := firstPosClone.Y; py <= t.lastPos.Y; py++ {
			for px := firstPosClone.X; px <= t.lastPos.X; px++ {
				CurrentFile.Selection = append(CurrentFile.Selection, IntVec2{px, py})
			}
		}
	}
}

// MouseUp is for mouse up events
func (t *SelectorTool) MouseUp(x, y int, button rl.MouseButton) {
	t.firstDown = false
	CurrentFile.DoingSelection = false
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
