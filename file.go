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

// HistoryType defines what kind of action happened
type HistoryType int

const (
	// PixelChangeHistoryType is used for pixel state changes
	PixelChangeHistoryType HistoryType = iota
	// AddLayerHistoryType is used when a layer is added
	AddLayerHistoryType
	// DeleteLayerHistoryType is used when a layer is deleted
	DeleteLayerHistoryType
)

// HistoryAction stores information about the action
type HistoryAction struct {
	HistoryType HistoryType // Defaults to PixelChangeHistoryType
	PixelState  map[IntVec2]PixelStateData
	LayerIndex  int
}

// DrawPixel draws a pixel. It records actions into history.
func (f *File) DrawPixel(x, y int, color rl.Color, saveToHistory bool) {

	// Set the pixel data in the current layer
	layer := f.GetCurrentLayer()
	if saveToHistory {
		if x >= 0 && y >= 0 && x < f.CanvasWidth && y < f.CanvasHeight {
			// Add old color to history
			rl.BeginTextureMode(layer.Canvas)
			rl.DrawPixel(x, y, color)
			rl.EndTextureMode()

			oldColor, ok := layer.PixelData[IntVec2{x, y}]
			if !ok {
				oldColor = rl.Transparent
			}

			// Prevent overwriting the old color with the new color since this
			// function is called every frame
			// Always draws to the last element of f.History since the
			// offset is removed automatically on mouse down
			if oldColor != color {
				ps := f.History[len(f.History)-1].PixelState[IntVec2{x, y}]
				ps.Current = color
				ps.Prev = oldColor
				f.History[len(f.History)-1].PixelState[IntVec2{x, y}] = ps
			}

			// Change pixel data to the new color
			layer.PixelData[IntVec2{x, y}] = color

			// Erase pixel by redrawing the entire canvas since drawing is
			//  additive only
			if color == rl.Transparent {
				f.DrawPixelDataToCanvas()
			}
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

	History           []HistoryAction
	HistoryMaxActions int
	historyOffset     int      // How many undos have been made
	deletedLayers     []*Layer // stack of layers, AddNewLayer destroys history chain

	Tools      []Tool // TODO Will be used by a Tool selector UI
	LeftTool   Tool
	RightTool  Tool
	LeftColor  rl.Color
	RightColor rl.Color
	// for preventing multiple event firing
	HasDoneMouseUpLeft  bool
	HasDoneMouseUpRight bool

	CanvasWidth, CanvasHeight, TileWidth, TileHeight int
}

// NewFile returns a pointer to a new File
func NewFile(canvasWidth, canvasHeight, tileWidth, tileHeight int) *File {
	f := &File{
		Filename: "filename",
		Layers: []*Layer{
			NewLayer(canvasWidth, canvasHeight, "background", rl.DarkGray, true),
			NewLayer(canvasWidth, canvasHeight, "hidden", rl.Transparent, true),
		},
		History:           make([]HistoryAction, 0, 50),
		HistoryMaxActions: 50, // TODO get from config
		deletedLayers:     make([]*Layer, 0, 10),

		LeftColor:  rl.Red,
		RightColor: rl.Blue,

		HasDoneMouseUpLeft:  true,
		HasDoneMouseUpRight: true,

		CanvasWidth:  canvasWidth, // TODO prompt
		CanvasHeight: canvasHeight,
		TileWidth:    tileWidth,
		TileHeight:   tileHeight,
	}
	f.LeftTool = NewPixelBrushTool("Pixel Brush L", false)
	f.RightTool = NewPixelBrushTool("Pixel Brush R", false)

	return f
}

// SetCurrentLayer sets the current layer
func (f *File) SetCurrentLayer(index int) {
	f.CurrentLayer = index
}

// GetCurrentLayer reutrns the current layer
func (f *File) GetCurrentLayer() *Layer {
	return f.Layers[f.CurrentLayer]
}

// AddNewLayer inserts a new layer
func (f *File) AddNewLayer() {
	newLayer := NewLayer(f.CanvasWidth, f.CanvasHeight, "new layer", rl.Transparent, true)
	f.Layers = append(f.Layers[:len(f.Layers)-1], newLayer, f.Layers[len(f.Layers)-1])
	f.SetCurrentLayer(len(f.Layers) - 2) // -2 bc temp layer is excluded

	f.AppendHistory(HistoryAction{AddLayerHistoryType, make(map[IntVec2]PixelStateData), f.CurrentLayer})
}

// AppendHistory inserts a new HistoryAction to f.History depending on the
// historyOffset
func (f *File) AppendHistory(action HistoryAction) {
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

// Undo usdoes an action
func (f *File) Undo() {
	if f.historyOffset < len(f.History) {
		f.historyOffset++
		index := len(f.History) - f.historyOffset
		history := f.History[index]

		log.Println("undo", history)

		switch history.HistoryType {
		case PixelChangeHistoryType:
			current := f.CurrentLayer
			f.SetCurrentLayer(history.LayerIndex)
			layer := f.GetCurrentLayer()
			rl.BeginTextureMode(layer.Canvas)
			rl.ClearBackground(rl.Transparent)
			shouldDraw := true
			for v, color := range layer.PixelData {
				shouldDraw = true
				for pos, psd := range history.PixelState {
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
					rl.DrawPixel(v.X, v.Y, color)
				}
			}
			rl.EndTextureMode()
			f.SetCurrentLayer(current)
		case AddLayerHistoryType:
			f.deletedLayers = append(f.deletedLayers, f.Layers[history.LayerIndex])
			f.Layers = append(f.Layers[:history.LayerIndex], f.Layers[history.LayerIndex+1:]...)
			f.SetCurrentLayer(history.LayerIndex - 1)
			LayersUIRebuildList() // TODO move UI stuff somewhere else
		}

	}
}

// Redo redoes an action
func (f *File) Redo() {
	if f.historyOffset > 0 {
		current := f.CurrentLayer
		index := len(f.History) - f.historyOffset
		history := f.History[index]

		log.Println("redo", history)

		switch history.HistoryType {
		case PixelChangeHistoryType:
			f.SetCurrentLayer(history.LayerIndex)
			layer := f.GetCurrentLayer()
			rl.BeginTextureMode(layer.Canvas)
			rl.ClearBackground(rl.Transparent)
			shouldDraw := true
			for v, color := range layer.PixelData {
				shouldDraw = true
				for pos, psd := range history.PixelState {
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
		case AddLayerHistoryType:
			layer := f.deletedLayers[len(f.deletedLayers)-1]
			f.deletedLayers = f.deletedLayers[:len(f.deletedLayers)-1]
			// TODO add to correct position on f.Layers
			f.Layers = append(f.Layers[:len(f.Layers)-1], layer, f.Layers[len(f.Layers)-1])
			LayersUIRebuildList()
		}

		f.historyOffset--
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
