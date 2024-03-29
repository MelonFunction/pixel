package main

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TODO rotate

// SelectorTool allows for a selection to be made
type SelectorTool struct {
	firstPos, lastPos IntVec2
	firstDown         bool
	mouseReleased     bool
	resizeSide        ResizeDirection
	// Should resize the original selection only
	oldWidth  int32
	oldHeight int32
	oldImg    *rl.Image
	// selection copied into another image for modification
	imgCopy            *rl.Image
	oldSelection       []rl.Color
	oldSelectionCopied bool
	// Cancels the selection if a click happens without drag
	firstDownTime time.Time
	name          string

	selectionFadeColor                     int32
	selectionFadeColorIncrease             int32 // increase by amount
	selectionFadeColorIncreasing           bool
	selectionFadeColorIncreaseTimeLast     time.Time
	selectionFadeColorIncreaseTimeInterval time.Duration // fps independence
}

// NewSelectorTool returns the selector tool
func NewSelectorTool(name string) *SelectorTool {
	return &SelectorTool{
		name:                                   name,
		mouseReleased:                          true,
		selectionFadeColor:                     128,
		selectionFadeColorIncrease:             8,
		selectionFadeColorIncreasing:           true,
		selectionFadeColorIncreaseTimeInterval: time.Second / 60,
	}
}

// MouseDown is for mouse down events
func (t *SelectorTool) MouseDown(x, y int32, button MouseButton) {
	// Only get the first position after mouse has just been clicked
	cl := CurrentFile.GetCurrentLayer()

	if t.firstDown == false {
		t.firstDown = true
		t.firstDownTime = time.Now()
		t.firstPos = IntVec2{x, y}

		// Resize selection
		x0, y0 := CurrentFile.SelectionBounds[0], CurrentFile.SelectionBounds[1]
		x1, y1 := CurrentFile.SelectionBounds[2], CurrentFile.SelectionBounds[3]
		if t.firstPos.Y >= y0-1 && t.firstPos.Y-1 <= y1 {
			if t.firstPos.X == x0-1 {
				t.resizeSide = ResizeCL
				CurrentFile.SelectionResizing = true
			}
			if t.firstPos.X-1 == x1 {
				t.resizeSide = ResizeCR
				CurrentFile.SelectionResizing = true
			}
		}
		if t.firstPos.X >= x0-1 && t.firstPos.X-1 <= x1 {
			if t.firstPos.Y == y0-1 {
				// TODO use bit operations
				if t.resizeSide == ResizeCL {
					t.resizeSide = ResizeTL
				} else if t.resizeSide == ResizeCR {
					t.resizeSide = ResizeTR
				} else {
					t.resizeSide = ResizeTC
				}
				CurrentFile.SelectionResizing = true
			}
			if t.firstPos.Y-1 == y1 {
				if t.resizeSide == ResizeCL {
					t.resizeSide = ResizeBL
				} else if t.resizeSide == ResizeCR {
					t.resizeSide = ResizeBR
				} else {
					t.resizeSide = ResizeBC
				}
				CurrentFile.SelectionResizing = true
			}
		}

		t.mouseReleased = false
	}

	t.lastPos = IntVec2{x, y}
	firstPosClone := t.firstPos

	// Bounds resizing
	if CurrentFile.SelectionResizing == true {
		if t.oldSelectionCopied == false {
			t.oldSelectionCopied = true
			t.oldSelection = CurrentFile.SelectionPixels

			CurrentFile.MoveSelection(0, 0)

			// Make an image from the selection
			t.oldWidth = int32(CurrentFile.SelectionBounds[2] - CurrentFile.SelectionBounds[0] + 1)
			t.oldHeight = int32(CurrentFile.SelectionBounds[3] - CurrentFile.SelectionBounds[1] + 1)
		}

		if len(t.oldSelection) == 0 {
			CurrentFile.SelectionResizing = false
			return
		}

		// Make a new image to modify using the old data
		tex := rl.LoadRenderTexture(t.oldWidth, t.oldHeight)
		rl.UpdateTexture(tex.Texture, t.oldSelection)
		t.oldImg = rl.LoadImageFromTexture(tex.Texture)
		rl.UnloadRenderTexture(tex)

		// Resize selection bounds
		// Selection bounds shifting logic so that the selection is flipped
		// without including the starting pixel
		bottom := func() {
			CurrentFile.SelectionBounds[1] = CurrentFile.OrigSelectionBounds[1]
			if CurrentFile.SelectionBounds[3] < CurrentFile.SelectionBounds[1]-1 {
				CurrentFile.SelectionBounds[1] = CurrentFile.OrigSelectionBounds[1] - 1
				CurrentFile.SelectionBounds[3] += 2
			}
		}
		top := func() {
			CurrentFile.SelectionBounds[3] = CurrentFile.OrigSelectionBounds[3]
			if CurrentFile.SelectionBounds[1] > CurrentFile.SelectionBounds[3]+1 {
				CurrentFile.SelectionBounds[3] = CurrentFile.OrigSelectionBounds[3] + 1
				CurrentFile.SelectionBounds[1] -= 2
			}
		}
		left := func() {
			CurrentFile.SelectionBounds[2] = CurrentFile.OrigSelectionBounds[2]
			if CurrentFile.SelectionBounds[0] > CurrentFile.SelectionBounds[2]+1 {
				CurrentFile.SelectionBounds[2] = CurrentFile.OrigSelectionBounds[2] + 1
				CurrentFile.SelectionBounds[0] -= 2
			}
		}
		right := func() {
			CurrentFile.SelectionBounds[0] = CurrentFile.OrigSelectionBounds[0]
			if CurrentFile.SelectionBounds[2] < CurrentFile.SelectionBounds[0]-1 {
				CurrentFile.SelectionBounds[0] = CurrentFile.OrigSelectionBounds[0] - 1
				CurrentFile.SelectionBounds[2] += 2
			}
		}
		switch t.resizeSide {
		case ResizeTL:
			CurrentFile.SelectionBounds[0] = t.lastPos.X + 1
			CurrentFile.SelectionBounds[1] = t.lastPos.Y + 1
			top()
			left()
		case ResizeTC:
			CurrentFile.SelectionBounds[1] = t.lastPos.Y + 1
			top()
		case ResizeTR:
			CurrentFile.SelectionBounds[2] = t.lastPos.X - 1
			CurrentFile.SelectionBounds[1] = t.lastPos.Y + 1
			top()
			right()
		case ResizeCL:
			CurrentFile.SelectionBounds[0] = t.lastPos.X + 1
			left()
		case ResizeCR:
			CurrentFile.SelectionBounds[2] = t.lastPos.X - 1
			right()
		case ResizeBL:
			CurrentFile.SelectionBounds[0] = t.lastPos.X + 1
			CurrentFile.SelectionBounds[3] = t.lastPos.Y - 1
			bottom()
			left()
		case ResizeBC:
			CurrentFile.SelectionBounds[3] = t.lastPos.Y - 1
			bottom()
		case ResizeBR:
			CurrentFile.SelectionBounds[2] = t.lastPos.X - 1
			CurrentFile.SelectionBounds[3] = t.lastPos.Y - 1
			bottom()
			right()
		}

		// Do the resize
		newWidth := CurrentFile.SelectionBounds[2] - CurrentFile.SelectionBounds[0] + 1
		newHeight := CurrentFile.SelectionBounds[3] - CurrentFile.SelectionBounds[1] + 1

		// Reset the selection
		// TODO it creates a lot of objects, not very efficient
		CurrentFile.Selection = make(map[IntVec2]rl.Color)

		// Handle selection flips
		if newWidth <= 0 {
			newWidth *= -1
			newWidth += 2
			rl.ImageFlipHorizontal(t.oldImg)
		}
		if newHeight <= 0 {
			newHeight *= -1
			newHeight += 2
			rl.ImageFlipVertical(t.oldImg)
		}

		if newWidth > 0 && newHeight > 0 {
			rl.ImageResizeNN(t.oldImg, newWidth, newHeight)
		}

		// Dump pixels back into the selection
		imgPixels := rl.LoadImageColors(t.oldImg)
		CurrentFile.SelectionPixels = imgPixels
		var count int
		minY := MinInt32(CurrentFile.SelectionBounds[1], CurrentFile.SelectionBounds[3])
		maxY := MaxInt32(CurrentFile.SelectionBounds[1], CurrentFile.SelectionBounds[3])
		minX := MinInt32(CurrentFile.SelectionBounds[0], CurrentFile.SelectionBounds[2])
		maxX := MaxInt32(CurrentFile.SelectionBounds[0], CurrentFile.SelectionBounds[2])
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				if count < len(imgPixels) {
					CurrentFile.Selection[IntVec2{x, y}] = imgPixels[count]
					count++
				}
			}
		}

		return
	}

	if t.lastPos.X < firstPosClone.X {
		t.lastPos.X, firstPosClone.X = firstPosClone.X, t.lastPos.X
	}
	if t.lastPos.Y < firstPosClone.Y {
		t.lastPos.Y, firstPosClone.Y = firstPosClone.Y, t.lastPos.Y
	}

	// Move the selection
	if CurrentFile.DoingSelection && t.firstPos.X > CurrentFile.SelectionBounds[0] && t.firstPos.X < CurrentFile.SelectionBounds[2] &&
		t.firstPos.Y > CurrentFile.SelectionBounds[1] && t.firstPos.Y < CurrentFile.SelectionBounds[3] {
		CurrentFile.MoveSelection(x-t.firstPos.X, y-t.firstPos.Y)
		t.firstPos.X = x
		t.firstPos.Y = y

		CurrentFile.OrigSelectionBounds[0] = CurrentFile.SelectionBounds[0]
		CurrentFile.OrigSelectionBounds[1] = CurrentFile.SelectionBounds[1]
		return
	}

	// Cancel selection if a click without a drag happens
	if t.firstPos.X == t.lastPos.X && t.firstPos.Y == t.lastPos.Y {
		if time.Now().Sub(t.firstDownTime) < time.Millisecond*100 {
			// Commit whatever was moving to wherever it ended up
			CurrentFile.CommitSelection()
			return
		}
	}

	// Reset the selection
	// TODO it creates a lot of objects, not very efficient
	CurrentFile.Selection = make(map[IntVec2]rl.Color)
	CurrentFile.SelectionPixels = make([]rl.Color, 0, (t.lastPos.X-firstPosClone.X)*(t.lastPos.Y-firstPosClone.Y))

	CurrentFile.SelectionBounds[0] = firstPosClone.X
	CurrentFile.SelectionBounds[1] = firstPosClone.Y
	CurrentFile.SelectionBounds[2] = t.lastPos.X
	CurrentFile.SelectionBounds[3] = t.lastPos.Y
	CurrentFile.OrigSelectionBounds[0] = CurrentFile.SelectionBounds[0]
	CurrentFile.OrigSelectionBounds[1] = CurrentFile.SelectionBounds[1]
	CurrentFile.OrigSelectionBounds[2] = CurrentFile.SelectionBounds[2]
	CurrentFile.OrigSelectionBounds[3] = CurrentFile.SelectionBounds[3]

	// Selection is being displayed on screen
	CurrentFile.DoingSelection = true

	for py := firstPosClone.Y; py <= t.lastPos.Y; py++ {
		for px := firstPosClone.X; px <= t.lastPos.X; px++ {
			pixel := cl.PixelData[IntVec2{px, py}]
			CurrentFile.Selection[IntVec2{px, py}] = pixel
			CurrentFile.SelectionPixels = append(CurrentFile.SelectionPixels, pixel)
		}
	}
}

// MouseUp is for mouse up events
func (t *SelectorTool) MouseUp(x, y int32, button MouseButton) {
	t.firstDown = false
	t.mouseReleased = true
	t.oldSelectionCopied = false
	CurrentFile.SelectionResizing = false
	t.resizeSide = ResizeNone

	if CurrentFile.SelectionBounds[2] < CurrentFile.SelectionBounds[0] {
		CurrentFile.SelectionBounds[2], CurrentFile.SelectionBounds[0] = CurrentFile.SelectionBounds[0], CurrentFile.SelectionBounds[2]
	}
	if CurrentFile.SelectionBounds[3] < CurrentFile.SelectionBounds[1] {
		CurrentFile.SelectionBounds[3], CurrentFile.SelectionBounds[1] = CurrentFile.SelectionBounds[1], CurrentFile.SelectionBounds[3]
	}
	CurrentFile.OrigSelectionBounds[0] = CurrentFile.SelectionBounds[0]
	CurrentFile.OrigSelectionBounds[1] = CurrentFile.SelectionBounds[1]
	CurrentFile.OrigSelectionBounds[2] = CurrentFile.SelectionBounds[2]
	CurrentFile.OrigSelectionBounds[3] = CurrentFile.SelectionBounds[3]
}

// DrawPreview is for drawing the preview
func (t *SelectorTool) DrawPreview(x, y int32) {
	rl.ClearBackground(rl.Blank)

	if CurrentFile.DoingSelection {
		// Draw the selected pixels
		for loc, color := range CurrentFile.Selection {
			rl.DrawPixel(loc.X, loc.Y, color)
		}
	}
}

// DrawUI is for drawing the UI
func (t *SelectorTool) DrawUI(camera rl.Camera2D) {
	if !CurrentFile.DoingSelection {
		return
	}
	pos := rl.GetWorldToScreen2D(rl.Vector2{X: float32(CurrentFile.SelectionBounds[0]) - float32(CurrentFile.CanvasWidth)/2, Y: float32(CurrentFile.SelectionBounds[1]) - float32(CurrentFile.CanvasHeight)/2}, camera)
	x := pos.X
	y := pos.Y
	w := float32(CurrentFile.SelectionBounds[2]-CurrentFile.SelectionBounds[0]+1) * camera.Zoom
	h := float32(CurrentFile.SelectionBounds[3]-CurrentFile.SelectionBounds[1]+1) * camera.Zoom

	if w <= 0 {
		x += w - 1*camera.Zoom
		w = w*-1 + 2*camera.Zoom
	}
	if h <= 0 {
		y += h - 1*camera.Zoom
		h = h*-1 + 2*camera.Zoom
	}

	if time.Now().Sub(t.selectionFadeColorIncreaseTimeLast) > t.selectionFadeColorIncreaseTimeInterval {
		t.selectionFadeColorIncreaseTimeLast = time.Now()

		if t.selectionFadeColorIncreasing {
			t.selectionFadeColor += t.selectionFadeColorIncrease
		} else {
			t.selectionFadeColor -= t.selectionFadeColorIncrease
		}

		if t.selectionFadeColor >= 255 {
			t.selectionFadeColorIncreasing = false
			t.selectionFadeColor = 255
		} else if t.selectionFadeColor <= 128 {
			t.selectionFadeColorIncreasing = true
			t.selectionFadeColor = 128
		}
	}

	// log.Println(t.selectionFadeColor)
	c := rl.NewColor(uint8(t.selectionFadeColor), uint8(t.selectionFadeColor), uint8(t.selectionFadeColor), 255)

	p := camera.Zoom                                                   // pixel size
	rl.DrawRectangleLinesEx(rl.NewRectangle(x, y, w, h), 4, c)         // main
	rl.DrawRectangleLinesEx(rl.NewRectangle(x-p, y-p, w+p*2, p), 2, c) // top
	rl.DrawRectangleLinesEx(rl.NewRectangle(x-p, y+h, w+p*2, p), 2, c) // bottom
	rl.DrawRectangleLinesEx(rl.NewRectangle(x-p, y-p, p, h+p*2), 2, c) // left
	rl.DrawRectangleLinesEx(rl.NewRectangle(x+w, y-p, p, h+p*2), 2, c) // right
}

func (t *SelectorTool) String() string {
	return t.name
}
