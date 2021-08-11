package main

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	rl "github.com/lachee/raylib-goplus/raylib"
)

// Tool is the interface for Tool elements
type Tool interface {
	// Used by every tool
	MouseDown(x, y int, button rl.MouseButton) // Called each frame the mouse is down
	MouseUp(x, y int, button rl.MouseButton)   // Called once, when the mouse button is released

	String() string

	// Takes the current mouse position. Called every frame the tool is
	// selected. Draw calls are drawn on the preview layer.
	DrawPreview(x, y int)
}

// HistoryLayerAction specifies the action which has been called upon the layer
type HistoryLayerAction int

// What HistoryLayer action has happened
const (
	HistoryLayerActionDelete HistoryLayerAction = iota
	HistoryLayerActionCreate
	HistoryLayerActionMoveUp
	HistoryLayerActionMoveDown
)

// HistoryLayer is for layer operations
type HistoryLayer struct {
	HistoryLayerAction
	LayerIndex int
}

// PixelStateData stores what the state was previously and currently
// Prev is used by undo and Current is used by redo
type PixelStateData struct {
	Prev, Current rl.Color
}

// HistoryPixel is for pixel operations
type HistoryPixel struct {
	PixelState map[IntVec2]PixelStateData
	LayerIndex int
}

// HistoryResize is for resize operations
type HistoryResize struct {
	// PrevLayerState is a slice consisting of all layer's PixelData
	PrevLayerState, CurrentLayerState []map[IntVec2]rl.Color
	// Used for calling Layer.Resize. ResizeDirection doesn't matter
	PrevWidth, PrevHeight       int
	CurrentWidth, CurrentHeight int
}

// DrawPixel draws a pixel. It records actions into history.
func (f *File) DrawPixel(x, y int, color rl.Color, saveToHistory bool) {
	// Set the pixel data in the current layer
	layer := f.GetCurrentLayer()
	if saveToHistory {
		if x >= 0 && y >= 0 && x < f.CanvasWidth && y < f.CanvasHeight {
			// Add old color to history
			oldColor, ok := layer.PixelData[IntVec2{x, y}]
			if !ok {
				oldColor = rl.Transparent
			}

			// TODO don't allow multiple opacity compressions per frame/event
			if color != rl.Transparent {
				color = BlendWithOpacity(oldColor, color)
			}

			// Prevent overwriting the old color with the new color since this
			// function is called every frame
			// Always draws to the last element of f.History since the
			// offset is removed automatically on mouse down
			if oldColor != color {
				latestHistoryInterface := f.History[len(f.History)-1]
				latestHistory, ok := latestHistoryInterface.(HistoryPixel)
				if ok {
					ps := latestHistory.PixelState[IntVec2{x, y}]
					ps.Current = color
					ps.Prev = oldColor
					latestHistory.PixelState[IntVec2{x, y}] = ps
				}
			}

			// Change pixel data to the new color
			layer.PixelData[IntVec2{x, y}] = color
			layer.Redraw()
		}

	}

}

// ClearBackground fills the initial PixelData
func (f *File) ClearBackground(color rl.Color) {
	rl.ClearBackground(color)

	layer := f.GetCurrentLayer()
	for x := 0; x < f.CanvasWidth; x++ {
		for y := 0; y < f.CanvasHeight; y++ {
			layer.PixelData[IntVec2{x, y}] = color
		}
	}
}

// FileSer contains only the fields that need to be serialized
type FileSer struct {
	DrawGrid bool

	Layers []*LayerSer
}
type LayerSer struct {
	Hidden        bool
	Name          string
	PixelData     map[IntVec2]rl.Color
	Width, Height int
}

// Animation contains data about an animation
type Animation struct {
	Name                 string
	FrameStart, FrameEnd int
	Timing               time.Duration // time between frames
}

// File contains all the methods and data required to alter a file
type File struct {
	// Save directory of the file
	PathDir  string
	Filename string

	Layers       []*Layer // The last one is for tool previews
	CurrentLayer int

	Animations       []*Animation
	CurrentAnimation int

	History           []interface{}
	HistoryMaxActions int
	historyOffset     int      // How many undos have been made
	deletedLayers     []*Layer // stack of layers, AddNewLayer destroys history chain

	LeftTool   Tool
	RightTool  Tool
	LeftColor  rl.Color
	RightColor rl.Color
	// For preventing multiple event firing
	HasDoneMouseUpLeft  bool
	HasDoneMouseUpRight bool

	// If grid should be drawn
	DrawGrid bool

	// Is selection happening currently
	DoingSelection bool
	// All of the affected pixels
	Selection map[IntVec2]rl.Color
	// Like above, but ordered
	SelectionPixels []rl.Color
	// Used for history appending, pixel overwriting/transparency logic
	// True after a selection has been made, false when nothing is selected
	SelectionMoving bool
	// SelectionResizing is true when the selection is being resized
	SelectionResizing bool
	// Bounds can be moved if dragged within this area
	SelectionBounds [4]int
	// To check if the selection was moved
	OrigSelectionBounds [4]int

	// CopiedSelection holds the selection when File.Copy is called
	CopiedSelection       map[IntVec2]rl.Color
	CopiedSelectionPixels []rl.Color
	// If the layer data should be moved or not
	IsSelectionPasted     bool
	CopiedSelectionBounds [4]int

	CurrentPalette int

	// Canvas and tile dimensions
	CanvasWidth, CanvasHeight, TileWidth, TileHeight int

	// for previewing what would happen if a resize occured
	DoingResize                                                                                          bool
	CanvasWidthResizePreview, CanvasHeightResizePreview, TileWidthResizePreview, TileHeightResizePreview int
	// direction of resize event
	CanvasDirectionResizePreview ResizeDirection
}

// NewFile returns a pointer to a new File
func NewFile(canvasWidth, canvasHeight, tileWidth, tileHeight int) *File {

	f := &File{
		Filename: "filename",
		Layers: []*Layer{
			NewLayer(canvasWidth, canvasHeight, "background", rl.Transparent, true),
			NewLayer(canvasWidth, canvasHeight, "hidden", rl.Transparent, true),
		},

		Animations: make([]*Animation, 0),

		History:           make([]interface{}, 0, 50),
		HistoryMaxActions: 50, // TODO get from config
		deletedLayers:     make([]*Layer, 0, 10),

		LeftColor:  rl.Red,
		RightColor: rl.Blue,

		HasDoneMouseUpLeft:  true,
		HasDoneMouseUpRight: true,

		DrawGrid: true,

		Selection: make(map[IntVec2]rl.Color),

		CanvasWidth:  canvasWidth,
		CanvasHeight: canvasHeight,
		TileWidth:    tileWidth,
		TileHeight:   tileHeight,

		CanvasWidthResizePreview:  canvasWidth,
		CanvasHeightResizePreview: canvasHeight,
		TileWidthResizePreview:    tileWidth,
		TileHeightResizePreview:   tileHeight,
	}
	// f.LeftTool = NewPixelBrushTool("Pixel Brush L", false)
	f.LeftTool = NewSpriteSelectorTool("Sprite Selector L")
	f.RightTool = NewPixelBrushTool("Pixel Brush R", false)

	return f
}

// ResizeDirection is used to specify which edge the resize operation applies to
type ResizeDirection int

// Resize directions
const (
	ResizeTL ResizeDirection = iota
	ResizeTC
	ResizeTR
	ResizeCL
	ResizeCC
	ResizeCR
	ResizeBL
	ResizeBC
	ResizeBR
)

// ResizeCanvas resizes the canvas from a specified edge
func (f *File) ResizeCanvas(width, height int, direction ResizeDirection) {
	prevLayerDatas := make([]map[IntVec2]rl.Color, 0, len(f.Layers))
	currentLayerDatas := make([]map[IntVec2]rl.Color, 0, len(f.Layers))

	for _, layer := range f.Layers {
		prevLayerDatas = append(prevLayerDatas, layer.PixelData)
		layer.Resize(width, height, direction)
		currentLayerDatas = append(currentLayerDatas, layer.PixelData)
	}

	f.AppendHistory(HistoryResize{prevLayerDatas, currentLayerDatas, f.CanvasWidth, f.CanvasHeight, width, height})
	f.CanvasWidth = width
	f.CanvasHeight = height

	LayersUIRebuildList()
}

// ResizeTileSize resizes the tile size
func (f *File) ResizeTileSize(width, height int) {
	f.TileWidth = width
	f.TileHeight = height
}

// DeleteSelection deletes the selection
func (f *File) DeleteSelection() {
	f.MoveSelection(0, 0)
	f.Selection = make(map[IntVec2]rl.Color)
}

// CancelSelection cancels the selection
func (f *File) CancelSelection() {
	f.Selection = make(map[IntVec2]rl.Color)
	f.SelectionMoving = false
	f.DoingSelection = false
}

// Copy the selection
func (f *File) Copy() {
	f.CopiedSelection = make(map[IntVec2]rl.Color)
	f.CopiedSelectionPixels = make([]rl.Color, 0, len(f.SelectionPixels))

	// Copy selection if there is one
	if len(f.Selection) > 0 {
		for v, c := range f.Selection {
			f.CopiedSelection[v] = c
		}
		for _, v := range f.SelectionPixels {
			f.CopiedSelectionPixels = append(f.CopiedSelectionPixels, v)
		}
		for i, v := range f.SelectionBounds {
			f.CopiedSelectionBounds[i] = v
		}
		return
	}

	// Otherwise copy the entire current layer
	cl := f.GetCurrentLayer()
	for v, c := range cl.PixelData {
		f.CopiedSelection[v] = c
	}
	f.CopiedSelectionBounds = [4]int{
		0,
		0,
		f.CanvasWidth - 1,
		f.CanvasHeight - 1,
	}

}

// Paste the selection
func (f *File) Paste() {
	f.CommitSelection()

	// Appends history
	f.SelectionMoving = false
	f.IsSelectionPasted = true
	f.MoveSelection(0, 0)
	f.DoingSelection = true

	f.Selection = make(map[IntVec2]rl.Color)
	for v, c := range f.CopiedSelection {
		f.Selection[v] = c
	}
	for _, v := range f.CopiedSelectionPixels {
		f.SelectionPixels = append(f.SelectionPixels, v)
	}

	for i, v := range f.CopiedSelectionBounds {
		f.SelectionBounds[i] = v
	}

	// TODO better way to switch tool
	if interactable, ok := toolSelector.GetInteractable(); ok {
		interactable.OnMouseUp(toolSelector, rl.MouseRightButton)
	}
}

// CommitSelection "stamps" the floating selection in place
func (f *File) CommitSelection() {
	f.IsSelectionPasted = false
	f.DoingSelection = false

	if f.SelectionMoving {
		f.SelectionMoving = false

		cl := f.GetCurrentLayer()

		// Alter PixelData and history
		for loc, color := range f.Selection {
			// Out of canvas bounds, ignore
			if !(loc.X >= 0 && loc.X < f.CanvasWidth && loc.Y >= 0 && loc.Y < f.CanvasHeight) {
				continue
			}

			latestHistoryInterface := f.History[len(f.History)-1]
			latestHistory, ok := latestHistoryInterface.(HistoryPixel)
			if ok {
				var currentColor rl.Color

				alreadyWritten, ok := latestHistory.PixelState[loc]
				if ok {
					currentColor = BlendWithOpacity(alreadyWritten.Current, color)
					// Overwrite the existing history
					alreadyWritten.Current = currentColor
					latestHistory.PixelState[loc] = alreadyWritten

				} else {
					currentColor = BlendWithOpacity(cl.PixelData[loc], color)
					ps := latestHistory.PixelState[loc]
					ps.Current = currentColor
					ps.Prev = cl.PixelData[loc]
					latestHistory.PixelState[loc] = ps

				}

				cl.PixelData[loc] = currentColor

			}
		}

		cl.Redraw()
	}

	// Reset the selection
	f.Selection = make(map[IntVec2]rl.Color)
	// Not important to reset this, but I'm doing it just because it feels right
	f.SelectionPixels = make([]rl.Color, 0, 0)

}

// MoveSelection moves the selection in the specified direction by one pixel
// dx and dy is how much the selection has moved
func (f *File) MoveSelection(dx, dy int) {
	cl := f.GetCurrentLayer()

	if len(f.Selection) > 0 {
		if !f.SelectionMoving {
			f.SelectionMoving = true

			f.AppendHistory(HistoryPixel{make(map[IntVec2]PixelStateData), CurrentFile.CurrentLayer})

			for loc := range f.Selection {
				// Alter history
				latestHistoryInterface := f.History[len(f.History)-1]
				latestHistory, ok := latestHistoryInterface.(HistoryPixel)
				if ok {
					ps := latestHistory.PixelState[loc]
					if !f.IsSelectionPasted {
						ps.Current = rl.Transparent
						ps.Prev = cl.PixelData[loc]
						latestHistory.PixelState[loc] = ps
					}
				}

				if !f.IsSelectionPasted {
					cl.PixelData[loc] = rl.Transparent
				}
			}
		}

		// Move selection
		CurrentFile.SelectionBounds[0] += dx
		CurrentFile.SelectionBounds[1] += dy
		CurrentFile.SelectionBounds[2] += dx
		CurrentFile.SelectionBounds[3] += dy
		newSelection := make(map[IntVec2]rl.Color)
		for loc, color := range f.Selection {
			newSelection[IntVec2{loc.X + dx, loc.Y + dy}] = color
		}
		f.Selection = newSelection

	}

	cl.Redraw()
}

// DeleteAnimation deletes an animation
func (f *File) DeleteAnimation(index int) error {
	return nil
}

// SetCurrentAnimation sets the current animation
func (f *File) SetCurrentAnimation(index int) {
	f.CurrentAnimation = index
}

// AddNewAnimation adds a new animation
func (f *File) AddNewAnimation() {
	f.Animations = append(f.Animations, &Animation{
		Name:       "New Animation",
		FrameStart: 0,
		FrameEnd:   0,
		Timing:     time.Second / 10,
	})
}

// SetCurrentLayer sets the current layer
func (f *File) SetCurrentLayer(index int) {
	f.CurrentLayer = index
}

// GetCurrentLayer returns the current layer
func (f *File) GetCurrentLayer() *Layer {
	return f.Layers[f.CurrentLayer]
}

// DeleteLayer deletes the layer.
// Won't delete anything if only one visible layer exists
// Sets the current layer to the top-most layer
func (f *File) DeleteLayer(index int, appendHistory bool) error {
	// TODO history
	if len(f.Layers) > 2 {
		f.deletedLayers = append(f.deletedLayers, f.Layers[index])
		f.Layers = append(f.Layers[:index], f.Layers[index+1:]...)

		if appendHistory {
			f.AppendHistory(HistoryLayer{HistoryLayerActionDelete, index})
			f.SetCurrentLayer(len(f.Layers) - 2)
		}

		return nil
	}

	return fmt.Errorf("Couldn't delete layer as it's the only one visible")
}

// RestoreLayer restores the last layer from f.deletedLayers to the position of
// index in f.Layers
func (f *File) RestoreLayer(index int) error {
	if len(f.deletedLayers) == 0 {
		return fmt.Errorf("No layers to restore")
	}

	f.Layers = append(f.Layers[:index], append([]*Layer{f.deletedLayers[len(f.deletedLayers)-1]}, f.Layers[index:]...)...)
	f.deletedLayers = append(f.deletedLayers[:len(f.deletedLayers)-1], f.deletedLayers[len(f.deletedLayers):]...)

	return nil
}

// AddNewLayer inserts a new layer
func (f *File) AddNewLayer() {
	newLayer := NewLayer(f.CanvasWidth, f.CanvasHeight, "new layer", rl.Transparent, true)
	f.Layers = append(f.Layers[:len(f.Layers)-1], newLayer, f.Layers[len(f.Layers)-1])
	f.SetCurrentLayer(len(f.Layers) - 2) // -2 bc temp layer is excluded

	f.AppendHistory(HistoryLayer{HistoryLayerActionCreate, f.CurrentLayer})
}

// MoveLayerUp moves the layer up
func (f *File) MoveLayerUp(index int) error {
	if index < len(f.Layers)-2 {
		toMove := f.Layers[index]
		f.Layers = append(f.Layers[:index], f.Layers[index+1:]...)
		f.Layers = append(f.Layers[:index], append([]*Layer{f.Layers[index], toMove}, f.Layers[index+1:]...)...)

		return nil
	}

	return fmt.Errorf("Couldn't move layer up")
}

// MoveLayerDown moves the layer down
func (f *File) MoveLayerDown(index int) error {
	// TODO history
	log.Println(index)
	if index > 0 {
		toMove := f.Layers[index]
		f.Layers = append(f.Layers[:index], f.Layers[index+1:]...)
		if index-1 == 0 {
			f.Layers = append([]*Layer{toMove}, append(f.Layers[:index], f.Layers[index:]...)...)
		} else {
			f.Layers = append(f.Layers[:index-1], append([]*Layer{toMove}, f.Layers[index-1:]...)...)
		}

		return nil
	}

	return fmt.Errorf("Couldn't move layer down")

}

// AppendHistory inserts a new history interface{} to f.History depending on the
// historyOffset
func (f *File) AppendHistory(action interface{}) {
	// Clear everything past the offset if a change has been made after undoing
	f.History = f.History[0 : len(f.History)-f.historyOffset]
	f.historyOffset = 0

	if len(f.History) >= f.HistoryMaxActions {
		f.History = append(f.History[len(f.History)-f.HistoryMaxActions+1:f.HistoryMaxActions], action)
	} else {
		f.History = append(f.History, action)
	}
}

// DrawPixelDataToCanvas redraws the canvas using the pixel data
// This is useful for removing pixels since DrawPixel is additive, meaning that
// a pixel can never be erased
func (f *File) DrawPixelDataToCanvas() {
	layer := f.GetCurrentLayer()
	rl.BeginTextureMode(layer.Canvas)
	rl.ClearBackground(rl.Transparent)
	for v, color := range layer.PixelData {
		rl.DrawPixel(v.X, v.Y, color)
	}
	rl.EndTextureMode()
}

// FlipHorizontal flips the layer horizontally, or flips the selection if anything
// is selected
func (f *File) FlipHorizontal() {
	latestHistory := HistoryPixel{make(map[IntVec2]PixelStateData), CurrentFile.CurrentLayer}

	sx, sy := 0, 0
	mx, my := f.CanvasWidth, f.CanvasHeight

	if f.DoingSelection {
		sx = f.SelectionBounds[0]
		sy = f.SelectionBounds[1]
		mx = (f.SelectionBounds[0] + f.SelectionBounds[2]) + 1
		my = f.SelectionBounds[3] + 1
	} else {
		// If selection is modified, it will be added to history on commit
		CurrentFile.AppendHistory(latestHistory)
	}

	// Swap the pixels over
	cl := f.GetCurrentLayer()
	wasSelectionMoving := f.SelectionMoving
	for y := sy; y < my; y++ {
		for x := sx; x < mx/2; x++ {
			lpos := IntVec2{x, y}
			rpos := IntVec2{mx - x - 1, y}

			lcur := cl.PixelData[lpos]
			rcur := cl.PixelData[rpos]

			// Update selection
			if f.DoingSelection {
				f.Selection[lpos], f.Selection[rpos] = f.Selection[rpos], f.Selection[lpos]
			} else {
				l := latestHistory.PixelState[lpos]
				l.Prev = lcur
				l.Current = rcur
				latestHistory.PixelState[lpos] = l

				r := latestHistory.PixelState[rpos]
				r.Prev = rcur
				r.Current = lcur
				latestHistory.PixelState[rpos] = r

				cl.PixelData[lpos] = rcur
				cl.PixelData[rpos] = lcur
			}

		}
	}

	if f.DoingSelection && wasSelectionMoving == false {
		// Allow CommitSelection to detect a change
		f.MoveSelection(0, 0)
	}

	cl.Redraw()
}

// FlipVertical flips the layer vertically, or flips the selection if anything
// is selected
func (f *File) FlipVertical() {
	latestHistory := HistoryPixel{make(map[IntVec2]PixelStateData), CurrentFile.CurrentLayer}

	sx, sy := 0, 0
	mx, my := f.CanvasWidth, f.CanvasHeight

	if f.DoingSelection {
		sx = f.SelectionBounds[0]
		sy = f.SelectionBounds[1]
		mx = f.SelectionBounds[2] + 1
		my = (f.SelectionBounds[1] + f.SelectionBounds[3]) + 1
	} else {
		// If selection is modified, it will be added to history on commit
		CurrentFile.AppendHistory(latestHistory)
	}

	// Swap the pixels over
	cl := f.GetCurrentLayer()
	wasSelectionMoving := f.SelectionMoving
	for x := sx; x < mx; x++ {
		for y := sy; y < my/2; y++ {
			lpos := IntVec2{x, y}
			rpos := IntVec2{x, my - y - 1}

			lcur := cl.PixelData[lpos]
			rcur := cl.PixelData[rpos]

			// Update selection
			if f.DoingSelection {
				f.Selection[lpos], f.Selection[rpos] = f.Selection[rpos], f.Selection[lpos]
			} else {
				l := latestHistory.PixelState[lpos]
				l.Prev = lcur
				l.Current = rcur
				latestHistory.PixelState[lpos] = l

				r := latestHistory.PixelState[rpos]
				r.Prev = rcur
				r.Current = lcur
				latestHistory.PixelState[rpos] = r

				cl.PixelData[lpos] = rcur
				cl.PixelData[rpos] = lcur
			}

		}
	}

	if f.DoingSelection && wasSelectionMoving == false {
		// Allow CommitSelection to detect a change
		f.MoveSelection(0, 0)
	}

	cl.Redraw()
}

// Undo undoes an action
func (f *File) Undo() {
	if f.historyOffset < len(f.History) {
		f.historyOffset++
		index := len(f.History) - f.historyOffset
		history := f.History[index]

		switch typed := history.(type) {
		case HistoryPixel:
			if f.DoingSelection {
				f.Selection = make(map[IntVec2]rl.Color)
				f.DoingSelection = false
				f.SelectionMoving = false
			}
			current := f.CurrentLayer
			f.SetCurrentLayer(typed.LayerIndex)
			layer := f.GetCurrentLayer()
			shouldDraw := true
			for v, color := range layer.PixelData {
				shouldDraw = true
				for pos, psd := range typed.PixelState {
					if v.X == pos.X && v.Y == pos.Y {
						// Update current color with previous color
						layer.PixelData[v] = psd.Prev
						// Overwrite
						color = psd.Prev
						if psd.Prev == rl.Transparent {
							shouldDraw = false
						}
					}
				}
				if shouldDraw {
					layer.PixelData[v] = color
				}
			}
			f.SetCurrentLayer(current)
			layer.Redraw()
		case HistoryLayer:
			switch typed.HistoryLayerAction {
			case HistoryLayerActionDelete:
				f.RestoreLayer(typed.LayerIndex)
			case HistoryLayerActionCreate:
				f.DeleteLayer(typed.LayerIndex, false)
			case HistoryLayerActionMoveUp:
				// TODO
			case HistoryLayerActionMoveDown:
			}
		case HistoryResize:
			f.CanvasWidthResizePreview = typed.PrevWidth
			f.CanvasHeightResizePreview = typed.PrevHeight
			f.CanvasWidth = typed.PrevWidth
			f.CanvasHeight = typed.PrevHeight
			for i, layer := range typed.PrevLayerState {
				f.Layers[i].PixelData = layer
				f.Layers[i].Resize(typed.PrevWidth, typed.PrevHeight, ResizeTL)
			}
		}

		LayersUIRebuildList()
	}
}

// Redo redoes an action
func (f *File) Redo() {
	if f.historyOffset > 0 {
		current := f.CurrentLayer
		index := len(f.History) - f.historyOffset
		history := f.History[index]

		switch typed := history.(type) {
		case HistoryPixel:
			f.SetCurrentLayer(typed.LayerIndex)
			layer := f.GetCurrentLayer()
			rl.BeginTextureMode(layer.Canvas)
			rl.ClearBackground(rl.Transparent)
			shouldDraw := true
			for v, color := range layer.PixelData {
				shouldDraw = true
				for pos, psd := range typed.PixelState {
					if v.X == pos.X && v.Y == pos.Y {
						layer.PixelData[v] = psd.Current
						color = psd.Current
						if psd.Current == rl.Transparent {
							shouldDraw = false
						}
					}
				}
				if shouldDraw {
					rl.DrawPixel(v.X, v.Y, color)
				}
			}
			rl.EndTextureMode()
			f.SetCurrentLayer(current)
		case HistoryLayer:
			switch typed.HistoryLayerAction {
			case HistoryLayerActionDelete:
				f.DeleteLayer(typed.LayerIndex, false)
			case HistoryLayerActionCreate:
				f.RestoreLayer(typed.LayerIndex)
			case HistoryLayerActionMoveUp:
			case HistoryLayerActionMoveDown:
			}
		case HistoryResize:
			f.CanvasWidthResizePreview = typed.CurrentWidth
			f.CanvasHeightResizePreview = typed.CurrentHeight
			f.CanvasWidth = typed.CurrentWidth
			f.CanvasHeight = typed.CurrentHeight
			for i, layer := range typed.CurrentLayerState {
				f.Layers[i].PixelData = layer
				f.Layers[i].Resize(typed.CurrentWidth, typed.CurrentHeight, ResizeTL)
			}
		}

		f.historyOffset--
		LayersUIRebuildList()
	}
}

// Destroy unloads each layer's canvas
func (f *File) Destroy() {
	for _, layer := range f.Layers {
		layer.Canvas.Unload()
	}

	for i, file := range Files {
		if file == f {
			Files = append(Files[:i], Files[i+1:]...)
			return
		}
	}
}

// Save the file into binary with gob
func (f *File) Save(path string) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("saving file", path)
	enc := gob.NewEncoder(file)

	gob.Register(rl.Color{})
	gob.Register(IntVec2{})

	fSer := &FileSer{
		DrawGrid: f.DrawGrid,
		Layers:   make([]*LayerSer, len(f.Layers)),
	}
	for l := range f.Layers {
		fSer.Layers[l] = &LayerSer{
			Name:      f.Layers[l].Name,
			Hidden:    f.Layers[l].Hidden,
			PixelData: f.Layers[l].PixelData,
			Width:     f.Layers[l].Width,
			Height:    f.Layers[l].Height,
		}
	}

	if err := enc.Encode(fSer); err != nil {
		log.Println(err)
	}

	if err := file.Close(); err != nil {
		log.Fatal(err)
	}

	// Change name in the tab
	spl := strings.Split(path, "/")
	f.Filename = spl[len(spl)-1]
	EditorsUIRebuild()
}

// Export the file into .png etc
// TODO remember last save path so resaving/exporting is faster
func (f *File) Export(path string) {
	// Create a colored image of the given width and height.
	img := image.NewNRGBA(image.Rect(0, 0, f.CanvasWidth, f.CanvasHeight))

	for _, layer := range f.Layers[:len(f.Layers)-1] {
		if !layer.Hidden {
			for pos, data := range layer.PixelData {
				// TODO layer blend modes
				if data.A != 0 {
					img.Set(pos.X, pos.Y, color.NRGBA{
						R: data.R,
						G: data.G,
						B: data.B,
						A: data.A,
					})
				}
			}
		}
	}

	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(file, img); err != nil {
		file.Close()
		log.Fatal(err)
	}

	if err := file.Close(); err != nil {
		log.Fatal(err)
	}

	// Change name in the tab
	spl := strings.Split(path, "/")
	f.Filename = spl[len(spl)-1]
	EditorsUIRebuild()
}

// Open a file
func Open(openPath string) *File {
	f := NewFile(64, 64, 8, 8)
	f.Filename = "Drawing"
	f.PathDir = path.Dir(openPath)

	fi, err := os.Stat(openPath)
	if err != nil {
		log.Println(err)
	}
	if fi.Mode().IsRegular() {
		reader, err := os.Open(openPath)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()

		switch filepath.Ext(openPath) {
		case ".pix":
			dec := gob.NewDecoder(reader)
			fileSer := &FileSer{}
			if err := dec.Decode(&fileSer); err != nil {
				log.Println(err)
			}

			f.Layers = make([]*Layer, len(fileSer.Layers))
			for i, layer := range fileSer.Layers {
				f.Layers[i] = &Layer{
					Name:           layer.Name,
					Hidden:         layer.Hidden,
					PixelData:      layer.PixelData,
					Width:          layer.Width,
					Height:         layer.Height,
					Canvas:         rl.LoadRenderTexture(layer.Width, layer.Height),
					hasInitialFill: true,
				}
				f.Layers[i].Redraw()
			}

			LayersUIRebuildList()

		case ".png":
			img, err := png.Decode(reader)
			if err != nil {
				log.Fatal(err)
			}

			f.CanvasWidth = img.Bounds().Max.X
			f.CanvasHeight = img.Bounds().Max.Y

			editedLayer := NewLayer(f.CanvasWidth, f.CanvasHeight, "background", rl.Transparent, false)

			rl.BeginTextureMode(editedLayer.Canvas)
			for x := 0; x < f.CanvasWidth; x++ {
				for y := 0; y < f.CanvasHeight; y++ {
					color := img.At(x, y)
					r, g, b, a := color.RGBA()
					rlColor := rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a))
					editedLayer.PixelData[IntVec2{x, y}] = rlColor
					rl.DrawPixel(x, y, rlColor)
				}
			}
			rl.EndTextureMode()

			f.Layers = []*Layer{
				editedLayer,
				NewLayer(f.CanvasWidth, f.CanvasHeight, "hidden", rl.Transparent, true),
			}

			spl := strings.Split(openPath, "/")
			f.Filename = spl[len(spl)-1]
		}
	}

	return f
}
