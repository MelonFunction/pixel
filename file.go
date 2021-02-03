package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	rl "github.com/lachee/raylib-goplus/raylib"
)

// Tool is the interface for Tool elements
type Tool interface {
	// Used by every tool
	MouseDown(x, y int) // Called each frame the mouse is down
	MouseUp(x, y int)   // Called once, when the mouse button is released

	// Used by drawing tools
	SetColor(rl.Color)
	GetColor() rl.Color

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

// HistoryAction stores information about the action
type HistoryAction struct {
	// Tool is the tool used for the action (GetName is used in history panel)
	PixelState map[IntVec2]PixelStateData
	LayerIndex int
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
		}

	}

}

// ClearBackground fills out the initial PixelData. Probably shouldn't be
// called every frame.
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
	Layers       []*Layer // The last one is for tool previews
	CurrentLayer int

	History           []HistoryAction
	HistoryMaxActions int
	historyOffset     int // How many undos have been made

	Tools               []Tool // TODO Will be used by a Tool selector UI
	LeftTool            Tool
	RightTool           Tool
	HasDoneMouseUpLeft  bool
	HasDoneMouseUpRight bool

	CanvasWidth, CanvasHeight, TileWidth, TileHeight int
}

// NewFile returns a pointer to a new File
func NewFile(canvasWidth, canvasHeight, tileWidth, tileHeight int) *File {
	f := &File{
		Layers: []*Layer{
			NewLayer(canvasWidth, canvasHeight, "background", rl.DarkGray),
			NewLayer(canvasWidth, canvasHeight, "layer 1", rl.Transparent),
			NewLayer(canvasWidth, canvasHeight, "layer 2", rl.Transparent),
			NewLayer(canvasWidth, canvasHeight, "hidden", rl.Transparent),
		},
		History:           make([]HistoryAction, 0, 5),
		HistoryMaxActions: 5, // TODO get from config

		HasDoneMouseUpLeft:  true,
		HasDoneMouseUpRight: true,

		CanvasWidth:  canvasWidth,
		CanvasHeight: canvasHeight,
		TileWidth:    tileWidth,
		TileHeight:   tileHeight,
	}
	f.LeftTool = NewPixelBrushTool(rl.Red, f, "Pixel Brush L")
	f.RightTool = NewPixelBrushTool(rl.Green, f, "Pixel Brush R")

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
	newLayer := NewLayer(f.CanvasWidth, f.CanvasHeight, "new layer", rl.Transparent)
	f.Layers = append(f.Layers[:len(f.Layers)-1], newLayer, f.Layers[len(f.Layers)-1])
	f.SetCurrentLayer(len(f.Layers) - 2) // -2 bc temp layer is excluded
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

// Undo usdoes an action
func (f *File) Undo() {
	if f.historyOffset < len(f.History) {
		current := f.CurrentLayer
		f.historyOffset++

		// TODO flatten the stuff below as it's mostly repeated
		f.SetCurrentLayer(f.History[len(f.History)-f.historyOffset].LayerIndex)
		layer := f.GetCurrentLayer()
		newCanvas := rl.LoadRenderTexture(f.CanvasWidth, f.CanvasHeight)
		rl.BeginTextureMode(newCanvas)
		shouldDraw := true
		for v, color := range layer.PixelData {
			shouldDraw = true
			for pos, psd := range f.History[len(f.History)-f.historyOffset].PixelState {
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
		layer.Canvas = newCanvas
		rl.EndTextureMode()

		f.SetCurrentLayer(current)
	}
}

// Redo redoes an action
func (f *File) Redo() {
	if f.historyOffset > 0 {
		current := f.CurrentLayer

		f.SetCurrentLayer(f.History[len(f.History)-f.historyOffset].LayerIndex)
		layer := f.GetCurrentLayer()
		newCanvas := rl.LoadRenderTexture(f.CanvasWidth, f.CanvasHeight)
		rl.BeginTextureMode(newCanvas)
		shouldDraw := true
		for v, color := range layer.PixelData {
			shouldDraw = true
			for pos, psd := range f.History[len(f.History)-f.historyOffset].PixelState {
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
		layer.Canvas = newCanvas
		rl.EndTextureMode()

		f.SetCurrentLayer(current)
		f.historyOffset--
	}
}

// Save the file into the custom editor format
func (f *File) Save() {

}

// Export the file into .png etc
func (f *File) Export() {
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

	// err if file exists
	_, err := os.Stat("image.png")
	if err == nil {
		log.Fatal(err)

		file, err := os.Create("image.png")
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
	}

}

// Destroy unloads each layer's canvas
func (f *File) Destroy() {
	for _, layer := range f.Layers {
		layer.Canvas.Unload()
	}
}
