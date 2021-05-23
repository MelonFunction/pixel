package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"strings"

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

// PixelStateData stores what the state was previously and currently
// Prev is used by undo and Current is used by redo
type PixelStateData struct {
	Prev, Current rl.Color
}

type HistoryLayer struct {
	WasDeleted bool
	LayerIndex int
}

type HistoryPixel struct {
	PixelState map[IntVec2]PixelStateData
	LayerIndex int
}

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

// File contains all the methods and data required to alter a file
type File struct {
	Filename     string
	Layers       []*Layer // The last one is for tool previews
	CurrentLayer int

	// History uses empty interfaces because I don't want to use nested structs
	// to defer the base type
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
	// Used for history appending, pixel overwriting/transparency logic
	// True after a selection has been made, false when nothing is selected
	SelectionMoving bool
	//Bounds can be moved if dragged within this area
	SelectionBounds [4]int
	// To check if the selection was moved
	OrigSelectionBounds [4]int

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
	f.LeftTool = NewPixelBrushTool("Pixel Brush L", false)
	f.RightTool = NewPixelBrushTool("Pixel Brush R", false)

	return f
}

type ResizeDirection int

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

func (f *File) Resize(width, height int, direction ResizeDirection) {
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

func (f *File) ResizeTileSize(width, height int) {
	f.TileWidth = width
	f.TileHeight = height
}

func (f *File) CommitSelection() {
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
}

// MoveSelection moves the selection in the specified direction by one pixel
func (f *File) MoveSelection(dx, dy int) {
	cl := f.GetCurrentLayer()

	if len(f.Selection) > 0 {
		if !f.SelectionMoving {
			f.SelectionMoving = true

			f.AppendHistory(HistoryPixel{make(map[IntVec2]PixelStateData), CurrentFile.CurrentLayer})

			for loc, color := range f.Selection {
				_ = loc
				_ = color
				// Alter history
				latestHistoryInterface := f.History[len(f.History)-1]
				latestHistory, ok := latestHistoryInterface.(HistoryPixel)
				if ok {
					ps := latestHistory.PixelState[loc]
					ps.Current = rl.Transparent
					ps.Prev = cl.PixelData[loc]
					latestHistory.PixelState[loc] = ps
				}

				cl.PixelData[loc] = rl.Transparent
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

// SetCurrentLayer sets the current layer
func (f *File) SetCurrentLayer(index int) {
	f.CurrentLayer = index
}

// GetCurrentLayer returns the current layer
func (f *File) GetCurrentLayer() *Layer {
	return f.Layers[f.CurrentLayer]
}

// AddNewLayer inserts a new layer
func (f *File) AddNewLayer() {
	newLayer := NewLayer(f.CanvasWidth, f.CanvasHeight, "new layer", rl.Transparent, true)
	f.Layers = append(f.Layers[:len(f.Layers)-1], newLayer, f.Layers[len(f.Layers)-1])
	f.SetCurrentLayer(len(f.Layers) - 2) // -2 bc temp layer is excluded

	f.AppendHistory(HistoryLayer{false, f.CurrentLayer})
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

// Undo usdoes an action
func (f *File) Undo() {
	if f.historyOffset < len(f.History) {
		f.historyOffset++
		index := len(f.History) - f.historyOffset
		history := f.History[index]

		switch typed := history.(type) {
		case HistoryPixel:
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
			if typed.WasDeleted {

			} else {
				f.deletedLayers = append(f.deletedLayers, f.Layers[typed.LayerIndex])
				f.Layers = append(f.Layers[:typed.LayerIndex], f.Layers[typed.LayerIndex+1:]...)
				f.SetCurrentLayer(typed.LayerIndex - 1)
			}
			LayersUIRebuildList()
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
			layer := f.deletedLayers[len(f.deletedLayers)-1]
			f.deletedLayers = f.deletedLayers[:len(f.deletedLayers)-1]
			// TODO add to correct position on f.Layers
			f.Layers = append(f.Layers[:len(f.Layers)-1], layer, f.Layers[len(f.Layers)-1])
			LayersUIRebuildList()
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

// Save the file into the custom editor format
func (f *File) Save(path string) {

}

// Export the file into .png etc
// TODO remember last save path so resaving/exporting is faster
func (f *File) Export(path string) {
	// Create a colored image of the given width and height.
	img := image.NewNRGBA(image.Rect(0, 0, f.CanvasWidth, f.CanvasHeight))

	for _, layer := range f.Layers[:len(f.Layers)-1] {
		log.Println(layer.Name)
		if !layer.Hidden {
			for pos, data := range layer.PixelData {
				if data.A == 255 {
					img.Set(pos.X, pos.Y, color.NRGBA{
						R: data.R,
						G: data.G,
						B: data.B,
						A: data.A,
					})
				} else {
					// TODO layer blend modes
					// Blend with existing depending on blend mode
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
func Open(path string) *File {
	reader, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	img, err := png.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	f := NewFile(img.Bounds().Max.X, img.Bounds().Max.Y, 8, 8)
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
	// Change name in the tab
	spl := strings.Split(path, "/")
	f.Filename = spl[len(spl)-1]
	return f
}
