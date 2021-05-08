package main

import (
	"time"

	rl "github.com/lachee/raylib-goplus/raylib"
)

// TODO resize
// TODO rotate

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
	cl := CurrentFile.GetCurrentLayer()
	if t.firstDown == false {
		t.firstDown = true
		t.firstDownTime = time.Now()
		t.firstPos = t.getClampedCoordinates(x, y)
	}

	t.lastPos = t.getClampedCoordinates(x, y)

	if t.firstPos.X > CurrentFile.SelectionBounds[0] && t.firstPos.X < CurrentFile.SelectionBounds[2] &&
		t.firstPos.Y > CurrentFile.SelectionBounds[1] && t.firstPos.Y < CurrentFile.SelectionBounds[3] {
		CurrentFile.MoveSelection(x-t.firstPos.X, y-t.firstPos.Y)
		t.firstPos.X = x
		t.firstPos.Y = y
		return
	}

	if t.firstPos.X == t.lastPos.X && t.firstPos.Y == t.lastPos.Y {
		// Cancel selection if a click without a drag happens
		if time.Now().Sub(t.firstDownTime) < time.Millisecond*100 {
			// Commit whatever was moving to wherever it ended up
			CurrentFile.CommitSelection()
			return
		}
	}

	// Reset the selection
	// TODO it creates a lot of objects, not very efficient
	CurrentFile.Selection = make(map[*IntVec2]rl.Color)

	firstPosClone := t.firstPos

	if t.lastPos.X < firstPosClone.X {
		t.lastPos.X, firstPosClone.X = firstPosClone.X, t.lastPos.X
	}
	if t.lastPos.Y < firstPosClone.Y {
		t.lastPos.Y, firstPosClone.Y = firstPosClone.Y, t.lastPos.Y
	}

	// TODO use comparison to make sure this is correct when using brush selector
	CurrentFile.SelectionBounds[0] = firstPosClone.X
	CurrentFile.SelectionBounds[1] = firstPosClone.Y
	CurrentFile.SelectionBounds[2] = t.lastPos.X
	CurrentFile.SelectionBounds[3] = t.lastPos.Y

	CurrentFile.OrigSelectionBounds = CurrentFile.SelectionBounds

	for py := firstPosClone.Y; py <= t.lastPos.Y; py++ {
		for px := firstPosClone.X; px <= t.lastPos.X; px++ {
			CurrentFile.Selection[&IntVec2{px, py}] = cl.PixelData[IntVec2{px, py}]
		}
	}
	// }

}

// MouseUp is for mouse up events
func (t *SelectorTool) MouseUp(x, y int, button rl.MouseButton) {
	t.firstDown = false
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

	// Draw selection overlay with handles after selection has finished
	if len(CurrentFile.Selection) > 0 {
		pa := IntVec2{CurrentFile.SelectionBounds[0], CurrentFile.SelectionBounds[1]}
		pb := IntVec2{CurrentFile.SelectionBounds[2], CurrentFile.SelectionBounds[3]}

		// top
		rl.DrawLineEx(
			rl.NewVector2(float32(pa.X), float32(pa.Y)),
			rl.NewVector2(float32(pb.X+1), float32(pa.Y)),
			1,
			rl.Color{255, 255, 255, 192})
		// bottom
		rl.DrawLineEx(
			rl.NewVector2(float32(pa.X), float32(pb.Y+2)),
			rl.NewVector2(float32(pb.X+1), float32(pb.Y+2)),
			1,
			rl.Color{255, 255, 255, 192})
		// left
		rl.DrawLineEx(
			rl.NewVector2(float32(pa.X-1), float32(pa.Y)),
			rl.NewVector2(float32(pa.X-1), float32(pb.Y+1)),
			1,
			rl.Color{255, 255, 255, 192})
		// right
		rl.DrawLineEx(
			rl.NewVector2(float32(pb.X+1), float32(pa.Y)),
			rl.NewVector2(float32(pb.X+1), float32(pb.Y+1)),
			1,
			rl.Color{255, 255, 255, 192})

		// Draw the selected pixels
		for loc, color := range CurrentFile.Selection {
			_ = color
			rl.DrawPixel(loc.X, loc.Y, color)
		}
	}
}

func (t *SelectorTool) String() string {
	return t.name
}
