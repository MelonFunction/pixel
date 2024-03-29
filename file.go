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

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Tool is the interface for Tool elements
type Tool interface {
	// Used by every tool
	MouseDown(x, y int32, button MouseButton) // Called each frame the mouse is down
	MouseUp(x, y int32, button MouseButton)   // Called once, when the mouse button is released

	String() string

	// Takes the current mouse position. Called every frame the tool is
	// selected. Draw calls are drawn on the preview layer.
	DrawPreview(x, y int32)

	// DrawUI is used to draw the UI for the current tool.
	// This doesn't draw to the preview layer.
	DrawUI(camera rl.Camera2D)
}

// HistoryLayerAction specifies the action which has been called upon the layer
type HistoryLayerAction int32

// What HistoryLayer action has happened
const (
	HistoryLayerActionDelete HistoryLayerAction = iota
	HistoryLayerActionCreate
	HistoryLayerActionMoveUp
	HistoryLayerActionMoveDown
)

//CompoundHistory is a group of history actions
type CompoundHistory struct {
	Actions []interface{}
}

// HistoryLayer is for layer operations
type HistoryLayer struct {
	HistoryLayerAction
	LayerIndex int32
}

// PixelStateData stores what the state was previously and currently
// Prev is used by undo and Current is used by redo
type PixelStateData struct {
	Prev, Current rl.Color
}

// HistoryPixel is for pixel operations
type HistoryPixel struct {
	PixelState map[IntVec2]PixelStateData
	LayerIndex int32
}

// HistoryResize is for resize operations
type HistoryResize struct {
	// PrevLayerState is a slice consisting of all layer's PixelData
	PrevLayerState, CurrentLayerState []map[IntVec2]rl.Color
	// Used for calling Layer.Resize. ResizeDirection doesn't matter
	PrevWidth, PrevHeight       int32
	CurrentWidth, CurrentHeight int32
}

// RedrawRenderLayer redraws the render layer
func (f *File) RedrawRenderLayer() {
	rl.BeginTextureMode(f.RenderLayer.Canvas)
	rl.ClearBackground(rl.Black)
	rl.BeginBlendMode(rl.BlendAlpha)
	for x := int32(0); x < f.CanvasWidth; x++ {
		for y := int32(0); y < f.CanvasHeight; y++ {
			color := rl.Blank
			loc := IntVec2{x, y}
			for _, layer := range f.Layers[:len(f.Layers)-1] {
				if !layer.Hidden {
					if layerColor, ok := layer.PixelData[loc]; ok {
						color = BlendWithOpacity(color, layerColor, layer.BlendMode)
					}
				}
			}
			f.RenderLayer.PixelData[loc] = color
			rl.DrawPixel(x, y, color)
		}
	}
	rl.EndBlendMode()
	rl.EndTextureMode()
}

// DrawPixel draws a pixel. It records actions into history.
// TODO replace all instances of accessing layer.PixelData with file.DrawPixel
func (f *File) DrawPixel(x, y int32, color rl.Color, layer *Layer) {
	// Set the pixel data in the current layer
	if x >= 0 && y >= 0 && x < f.CanvasWidth && y < f.CanvasHeight {
		loc := IntVec2{x, y}

		// Add old color to history
		oldColor, ok := layer.PixelData[loc]
		if !ok {
			oldColor = rl.Blank
		}

		// Blend color on passed layer
		if color != rl.Blank {
			color = BlendWithOpacity(oldColor, color, layer.BlendMode)
		}
		layer.PixelData[loc] = color

		// Prevent overwriting the old color with the new color since this function is called every frame
		// Always draws to the last element of f.History since the offset is removed automatically on mouse down
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

		// Draw to passed layer
		rl.BeginTextureMode(layer.Canvas)
		if color == rl.Blank {
			rl.DrawPixel(x, y, rl.Black)
		} else {
			rl.BeginBlendMode(layer.BlendMode)
			rl.DrawPixel(x, y, rl.Black)
			rl.DrawPixel(x, y, color)
			rl.EndBlendMode()
		}
		rl.EndTextureMode()

		// Draw to render layer
		rl.BeginTextureMode(f.RenderLayer.Canvas)

		// Erase current pixel color
		rl.BeginBlendMode(rl.BlendSubtractColors)
		rl.DrawPixel(x, y, f.RenderLayer.PixelData[loc])
		rl.EndBlendMode()

		rl.BeginBlendMode(rl.BlendAlpha)
		nc := rl.Blank
		for _, layer := range f.Layers[:len(f.Layers)-1] {
			if !layer.Hidden {
				if layerColor, ok := layer.PixelData[loc]; ok {
					nc = BlendWithOpacity(nc, layerColor, layer.BlendMode)
				}
			}
		}
		f.RenderLayer.PixelData[loc] = nc
		rl.DrawPixel(x, y, rl.Black)
		rl.DrawPixel(x, y, nc)
		rl.EndBlendMode()
		rl.EndTextureMode()

	}
}

// ClearBackground fills the initial PixelData
func (f *File) ClearBackground(color rl.Color) {
	rl.ClearBackground(color)

	layer := f.GetCurrentLayer()
	for x := int32(0); x < f.CanvasWidth; x++ {
		for y := int32(0); y < f.CanvasHeight; y++ {
			layer.PixelData[IntVec2{x, y}] = color
		}
	}
}

// FileSer contains only the fields that need to be serialized
type FileSer struct {
	DrawGrid                                         bool
	CanvasWidth, CanvasHeight, TileWidth, TileHeight int32

	Layers     []*LayerSer
	Animations []*AnimationSer
}

// LayerSer contains only the fields that need to be serialized
type LayerSer struct {
	Hidden        bool
	Name          string
	PixelData     map[IntVec2]rl.Color
	Width, Height int32
}

// AnimationSer contains only the fields that need to be serialized
type AnimationSer struct {
	Name                 string
	FrameStart, FrameEnd int32
	Timing               float32
}

// Animation contains data about an animation
type Animation struct {
	Name                 string
	FrameStart, FrameEnd int32
	Timing               float32 // time between frames
}

// File contains all the methods and data required to alter a file
type File struct {
	// Save directory of the file
	PathDir string
	// Save location of the file
	FileDir  string
	Filename string

	// FileChanged is true if a change has been made since saving
	FileChanged bool

	Layers       []*Layer // The last one is for tool previews
	RenderLayer  *Layer   // Blends all layers and renders only this layer
	CurrentLayer int32

	Animations       []*Animation
	CurrentAnimation int32

	History           []interface{}
	HistoryMaxActions int32
	historyOffset     int32    // How many undos have been made
	deletedLayers     []*Layer // stack of layers, AddNewLayer destroys history chain

	// For preventing multiple event firing
	HasDoneMouseUpLeft  bool
	HasDoneMouseUpRight bool

	// If grid should be drawn
	DrawGrid bool

	// Used by system_file.go
	FileCameraTarget rl.Vector2 // temp storage for calculations
	FileCamera       rl.Camera2D

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
	SelectionBounds [4]int32
	// To check if the selection was moved
	OrigSelectionBounds [4]int32
	// True if paste event has just happened
	IsSelectionPasted bool

	CurrentPalette int32

	// Canvas and tile dimensions
	CanvasWidth, CanvasHeight, TileWidth, TileHeight int32

	// for previewing what would happen if a resize occured
	DoingResize                                                                                          bool
	CanvasWidthResizePreview, CanvasHeightResizePreview, TileWidthResizePreview, TileHeightResizePreview int32
	// direction of resize event
	CanvasDirectionResizePreview ResizeDirection

	// Hacky way to handle events afer all values have been calculated.
	// There are probably race conditions here since the execution is delayed. If the file is switched before the
	// function is called, unexpected results will likely happen
	// It's only used by UIControlSystem.process to run the current palette color indicator's OnScroll function after
	// offset values  have been calculated
	RenderSystemRenderCallback chan func()
}

// NewFile returns a pointer to a new File
func NewFile(canvasWidth, canvasHeight, tileWidth, tileHeight int32) *File {

	var scaleRatio = 64.0 / float32(canvasHeight)

	pathDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	f := &File{
		PathDir:  pathDir,
		Filename: "filename",
		Layers: []*Layer{
			NewLayer(canvasWidth, canvasHeight, "background", rl.Blank, true),
			NewLayer(canvasWidth, canvasHeight, "hidden", rl.Blank, true),
		},
		RenderLayer: NewLayer(canvasWidth, canvasHeight, "render", rl.Blank, true),

		FileChanged: false,

		Animations: make([]*Animation, 0),

		History:           make([]interface{}, 0, 50),
		HistoryMaxActions: 500, // TODO get from config
		deletedLayers:     make([]*Layer, 0, 10),

		HasDoneMouseUpLeft:  true,
		HasDoneMouseUpRight: true,

		DrawGrid: canvasHeight <= 64, // don't draw the grid for anything bigger than default size

		FileCamera: rl.Camera2D{Zoom: 12.0 * scaleRatio,
			Offset: rl.NewVector2(
				float32(rl.GetScreenWidth())/2,
				float32(rl.GetScreenHeight())/2,
			)},

		Selection: make(map[IntVec2]rl.Color),

		CanvasWidth:  canvasWidth,
		CanvasHeight: canvasHeight,
		TileWidth:    tileWidth,
		TileHeight:   tileHeight,

		CanvasWidthResizePreview:  canvasWidth,
		CanvasHeightResizePreview: canvasHeight,
		TileWidthResizePreview:    tileWidth,
		TileHeightResizePreview:   tileHeight,

		RenderSystemRenderCallback: make(chan func(), 10),
	}

	// f.Layers[0].BlendMode = rl.BlendAddColors
	// f.RenderLayer.BlendMode = rl.BlendAddColors

	return f
}

// ResizeDirection is used to specify which edge the resize operation applies to
type ResizeDirection int32

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
	ResizeNone
)

// ResizeCanvas resizes the canvas from a specified edge
func (f *File) ResizeCanvas(width, height int32, direction ResizeDirection) {
	prevLayerDatas := make([]map[IntVec2]rl.Color, 0, len(f.Layers))
	currentLayerDatas := make([]map[IntVec2]rl.Color, 0, len(f.Layers))

	for _, layer := range f.Layers {
		prevLayerDatas = append(prevLayerDatas, layer.PixelData)
		layer.Resize(width, height, direction)
		currentLayerDatas = append(currentLayerDatas, layer.PixelData)
	}
	f.RenderLayer.Resize(width, height, direction)

	f.AppendHistory(HistoryResize{prevLayerDatas, currentLayerDatas, f.CanvasWidth, f.CanvasHeight, width, height})
	f.CanvasWidth = width
	f.CanvasHeight = height

	f.RedrawRenderLayer()
	LayersUIRebuildList()
}

// ResizeTileSize resizes the tile size
func (f *File) ResizeTileSize(width, height int32) {
	f.RedrawRenderLayer()
	f.TileWidth = width
	f.TileHeight = height
}

// DeleteSelection deletes the selection
func (f *File) DeleteSelection() {
	f.RedrawRenderLayer()
	f.MoveSelection(0, 0)
	f.Selection = make(map[IntVec2]rl.Color)
	f.SelectionPixels = make([]rl.Color, 0)
	f.CommitSelection()
}

// CancelSelection cancels the selection
func (f *File) CancelSelection() {
	f.RedrawRenderLayer()
	f.Selection = make(map[IntVec2]rl.Color)
	f.SelectionMoving = false
	f.DoingSelection = false
}

// Copy the selection
func (f *File) Copy() {
	CopiedSelection = make(map[IntVec2]rl.Color)
	CopiedSelectionPixels = make([]rl.Color, 0, len(f.SelectionPixels))

	// Copy selection if there is one
	if len(f.Selection) > 0 {
		for v, c := range f.Selection {
			CopiedSelection[v] = c
		}
		for _, v := range f.SelectionPixels {
			CopiedSelectionPixels = append(CopiedSelectionPixels, v)
		}
		for i, v := range f.SelectionBounds {
			CopiedSelectionBounds[i] = v
		}
		return
	}

	// Otherwise copy the entire current layer
	cl := f.GetCurrentLayer()
	for v, c := range cl.PixelData {
		CopiedSelection[v] = c
	}
	CopiedSelectionBounds = [4]int32{
		0,
		0,
		f.CanvasWidth - 1,
		f.CanvasHeight - 1,
	}

}

// Paste the selection
func (f *File) Paste() {
	// Appends history
	f.SelectionMoving = false
	f.IsSelectionPasted = true
	f.DoingSelection = true

	f.Selection = make(map[IntVec2]rl.Color)
	for v, c := range CopiedSelection {
		f.Selection[v] = c
	}
	for _, v := range CopiedSelectionPixels {
		f.SelectionPixels = append(f.SelectionPixels, v)
	}

	for i, v := range CopiedSelectionBounds {
		f.SelectionBounds[i] = v
	}

	// Use this to copy selection to buffer
	f.MoveSelection(0, 0)

	// TODO better way to switch tool
	if interactable, ok := toolSelector.GetInteractable(); ok {
		interactable.OnMouseUp(toolSelector, rl.MouseRightButton)
	}

	f.RedrawRenderLayer()
}

// MoveSelection moves the selection in the specified direction by one pixel
// dx and dy is how much the selection has moved
func (f *File) MoveSelection(dx, dy int32) {
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
						ps.Current = rl.Blank
						ps.Prev = cl.PixelData[loc]
						latestHistory.PixelState[loc] = ps
					}
				}

				if !f.IsSelectionPasted {
					cl.PixelData[loc] = rl.Blank
				}
			}
		}

		// Move selection
		f.SelectionBounds[0] += dx
		f.SelectionBounds[1] += dy
		f.SelectionBounds[2] += dx
		f.SelectionBounds[3] += dy
		f.OrigSelectionBounds[0] = f.SelectionBounds[0]
		f.OrigSelectionBounds[1] = f.SelectionBounds[1]
		f.OrigSelectionBounds[2] = f.SelectionBounds[2]
		f.OrigSelectionBounds[3] = f.SelectionBounds[3]

		newSelection := make(map[IntVec2]rl.Color)
		for loc, color := range f.Selection {
			newSelection[IntVec2{loc.X + dx, loc.Y + dy}] = color
		}
		f.Selection = newSelection

	}

	cl.Redraw()
	f.RedrawRenderLayer()
}

// CommitSelection "stamps" the floating selection in place
func (f *File) CommitSelection() {
	f.IsSelectionPasted = false
	f.DoingSelection = false

	if f.SelectionMoving {
		f.SelectionMoving = false

		if len(f.History) <= 0 {
			return
		}

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
					currentColor = BlendWithOpacity(alreadyWritten.Current, color, cl.BlendMode)
					// Overwrite the existing history
					alreadyWritten.Current = currentColor
					latestHistory.PixelState[loc] = alreadyWritten

				} else {
					currentColor = BlendWithOpacity(cl.PixelData[loc], color, cl.BlendMode)
					ps := latestHistory.PixelState[loc]
					ps.Current = currentColor
					ps.Prev = cl.PixelData[loc]
					latestHistory.PixelState[loc] = ps

				}

				cl.PixelData[loc] = currentColor

			}
		}

		cl.Redraw()
		f.RedrawRenderLayer()
	}

	// Reset the selection
	f.Selection = make(map[IntVec2]rl.Color)
	f.SelectionPixels = make([]rl.Color, 0, 0)
}

// DeleteAnimation deletes an animation
func (f *File) DeleteAnimation(index int32) error {
	if index-1 >= int32(len(f.Animations)) {
		return fmt.Errorf("Animation not in range")
	}

	f.Animations = append(f.Animations[:index], f.Animations[index+1:]...)
	// set animation to last
	f.CurrentAnimation = int32(len(f.Animations) - 1)

	return nil
}

// SetCurrentAnimation sets the current animation
func (f *File) SetCurrentAnimation(index int32) {
	f.CurrentAnimation = index
}

// GetCurrentAnimation gets the current animation
func (f *File) GetCurrentAnimation() *Animation {
	if len(f.Animations) == 0 {
		return nil
	}
	return f.Animations[f.CurrentAnimation]
}

// GetAnimation gets the animation at the specified index
func (f *File) GetAnimation(index int32) (*Animation, error) {
	if index-1 >= int32(len(f.Animations)) {
		return nil, fmt.Errorf("Animation not in range")
	}
	return f.Animations[index], nil
}

// AddNewAnimation adds a new animation
func (f *File) AddNewAnimation() {
	f.Animations = append(f.Animations, &Animation{
		Name:       fmt.Sprintf("Anim %d", len(f.Animations)),
		FrameStart: 0,
		FrameEnd:   0,
		Timing:     5.0, // 5 fps
	})
}

// SetAnimationFrames sets the current animation's frames
func (f *File) SetAnimationFrames(index, firstSprite, lastSprite int32) {
	anim, err := f.GetAnimation(index)
	if err != nil {
		log.Println(err)
		return
	}
	anim.FrameStart = firstSprite
	anim.FrameEnd = lastSprite
}

// SetCurrentAnimationTiming sets the current animation's timing
// The argument is the frames per second
func (f *File) SetCurrentAnimationTiming(timing float32) {
	anim := f.GetCurrentAnimation()
	if anim != nil {
		anim.Timing = timing
	}
}

// SetAnimationName sets the current animation's name
func (f *File) SetAnimationName(index int32, name string) {
	anim, err := f.GetAnimation(index)
	if err != nil {
		log.Println(err)
		return
	}
	anim.Name = name
}

// SetCurrentLayer sets the current layer
func (f *File) SetCurrentLayer(index int32) {
	f.CurrentLayer = index
}

// GetCurrentLayer returns the current layer
func (f *File) GetCurrentLayer() *Layer {
	return f.Layers[f.CurrentLayer]
}

// DeleteLayer deletes the layer.
// Won't delete anything if only one visible layer exists
// Sets the current layer to the top-most layer
func (f *File) DeleteLayer(index int32, appendHistory bool) error {
	if len(f.Layers) > 2 {
		f.deletedLayers = append(f.deletedLayers, f.Layers[index])
		f.Layers = append(f.Layers[:index], f.Layers[index+1:]...)

		if appendHistory {
			f.AppendHistory(HistoryLayer{HistoryLayerActionDelete, index})
		}

		if f.CurrentLayer > int32(len(f.Layers)-2) {
			f.SetCurrentLayer(int32(len(f.Layers) - 2))
		}

		return nil
	}

	return fmt.Errorf("Couldn't delete layer as it's the only one visible")
}

// RestoreLayer restores the last layer from f.deletedLayers to the position of
// index in f.Layers
func (f *File) RestoreLayer(index int32) error {
	if len(f.deletedLayers) == 0 {
		return fmt.Errorf("No layers to restore")
	}

	f.Layers = append(
		f.Layers[:index],
		append(
			[]*Layer{f.deletedLayers[len(f.deletedLayers)-1]},
			f.Layers[index:]...)...)
	f.deletedLayers = append(
		f.deletedLayers[:len(f.deletedLayers)-1],
		f.deletedLayers[len(f.deletedLayers):]...)

	f.RedrawRenderLayer()
	return nil
}

// MergeLayerDown merges the layer with the one below
func (f *File) MergeLayerDown(index int32) error {
	if len(f.Layers) <= 2 {
		return fmt.Errorf("Couldn't merge layer down: Not enough layers")
	}
	if index == 0 {
		return fmt.Errorf("Couldn't merge layer down: Can't merge lowest layer")
	}

	// old layer pixel state
	historyPixel := HistoryPixel{make(map[IntVec2]PixelStateData), index - 1}
	from := f.Layers[index]
	to := f.Layers[index-1]
	for loc, color := range from.PixelData {
		hist := historyPixel.PixelState[loc]
		hist.Prev = to.PixelData[loc]
		newColor := BlendWithOpacity(to.PixelData[loc], color, from.BlendMode)
		to.PixelData[loc] = newColor
		hist.Current = newColor

		// Save back into the map
		historyPixel.PixelState[loc] = hist

		if color != rl.Blank && color != to.PixelData[loc] {
		}
	}
	to.Redraw()

	if err := f.DeleteLayer(index, false); err != nil {
		return err
	}

	comp := CompoundHistory{
		Actions: []interface{}{
			historyPixel,
			HistoryLayer{HistoryLayerActionDelete, index},
		},
	}
	f.AppendHistory(comp)

	f.RedrawRenderLayer()
	return nil
}

// AddNewLayer inserts a new layer
func (f *File) AddNewLayer() {
	newLayer := NewLayer(f.CanvasWidth, f.CanvasHeight, "new layer", rl.Blank, true)
	f.Layers = append(f.Layers[:len(f.Layers)-1], newLayer, f.Layers[len(f.Layers)-1])
	f.SetCurrentLayer(int32(len(f.Layers) - 2)) // -2 bc temp layer is excluded

	f.AppendHistory(HistoryLayer{HistoryLayerActionCreate, f.CurrentLayer})
	f.RedrawRenderLayer()
}

// MoveLayerUp moves the layer up
func (f *File) MoveLayerUp(index int32, appendHistory bool) error {
	if index < int32(len(f.Layers)-2) {
		toMove := f.Layers[index]
		f.Layers = append(f.Layers[:index], f.Layers[index+1:]...)
		f.Layers = append(f.Layers[:index], append([]*Layer{f.Layers[index], toMove}, f.Layers[index+1:]...)...)

		if appendHistory {
			f.AppendHistory(HistoryLayer{HistoryLayerActionMoveUp, index})
		}
		f.RedrawRenderLayer()
		return nil
	}

	return fmt.Errorf("Couldn't move layer up")
}

// MoveLayerDown moves the layer down
func (f *File) MoveLayerDown(index int32, appendHistory bool) error {
	if index > 0 {
		toMove := f.Layers[index]
		f.Layers = append(f.Layers[:index], f.Layers[index+1:]...)
		if index-1 == 0 {
			f.Layers = append([]*Layer{toMove}, append(f.Layers[:index], f.Layers[index:]...)...)
		} else {
			f.Layers = append(f.Layers[:index-1], append([]*Layer{toMove}, f.Layers[index-1:]...)...)
		}

		if appendHistory {
			f.AppendHistory(HistoryLayer{HistoryLayerActionMoveDown, index})
		}
		f.RedrawRenderLayer()
		return nil
	}

	return fmt.Errorf("Couldn't move layer down")

}

// AppendHistory inserts a new history interface{} to f.History depending on the
// historyOffset
func (f *File) AppendHistory(action interface{}) {
	f.FileChanged = true
	// Clear everything past the offset if a change has been made after undoing
	f.History = f.History[0 : int32(len(f.History))-f.historyOffset]
	f.historyOffset = 0

	if int32(len(f.History)) >= f.HistoryMaxActions {
		f.History = append(f.History[int32(len(f.History))-f.HistoryMaxActions+1:f.HistoryMaxActions], action)
	} else {
		f.History = append(f.History, action)
	}

	EditorsUIRebuild()
}

// DrawPixelDataToCanvas redraws the canvas using the pixel data
// This is useful for removing pixels since DrawPixel is additive, meaning that
// a pixel can never be erased
func (f *File) DrawPixelDataToCanvas() {
	layer := f.GetCurrentLayer()
	rl.BeginTextureMode(layer.Canvas)
	rl.ClearBackground(rl.Blank)
	for v, color := range layer.PixelData {
		rl.DrawPixel(v.X, v.Y, color)
	}
	rl.EndTextureMode()
}

// Outline draws the left color around any non-transparent pixels (and is
// restricted to the selection)
// Will only outline pixels on the current layer. Make sure to merge layers if
// sprite is composed of multiple parts
func (f *File) Outline() {
	var sx, sy int32 = 0, 0
	mx, my := f.CanvasWidth, f.CanvasHeight

	latestHistory := HistoryPixel{make(map[IntVec2]PixelStateData), CurrentFile.CurrentLayer}
	if f.DoingSelection {
		// latestHistory is essentially ignored and whatever is in the selection
		// is accounted for by f.MoveSelection
		sx = f.SelectionBounds[0]
		sy = f.SelectionBounds[1]
		mx = f.SelectionBounds[2] + 1
		my = f.SelectionBounds[3] + 1
	} else {
		// New history
		CurrentFile.AppendHistory(latestHistory)
	}

	cl := f.GetCurrentLayer()
	pixelLocations := make([]IntVec2, 0, 0)

	for y := sy; y < my; y++ {
		for x := sx; x < mx; x++ {
			currentPos := IntVec2{x, y}
			leftPos := IntVec2{x - 1, y}
			rightPos := IntVec2{x + 1, y}
			abovePos := IntVec2{x, y - 1}
			belowPos := IntVec2{x, y + 1}

			// Change where the pixels are sampled from if there is a selection
			pixelSource := cl.PixelData
			if f.DoingSelection {
				pixelSource = f.Selection
			}

			if pixelSource[currentPos] != rl.Blank {
				if pixelSource[leftPos] == rl.Blank {
					pixelLocations = append(pixelLocations, leftPos)
				}
				if pixelSource[rightPos] == rl.Blank {
					pixelLocations = append(pixelLocations, rightPos)
				}
				if pixelSource[abovePos] == rl.Blank {
					pixelLocations = append(pixelLocations, abovePos)
				}
				if pixelSource[belowPos] == rl.Blank {
					pixelLocations = append(pixelLocations, belowPos)
				}
			}
		}
	}

	for _, loc := range pixelLocations {
		l := latestHistory.PixelState[loc]
		l.Prev = rl.Blank // Only replacing transparent pixels
		l.Current = LeftColor
		latestHistory.PixelState[loc] = l

		if f.DoingSelection {
			f.Selection[loc] = LeftColor
		} else {
			cl.PixelData[loc] = LeftColor
		}
	}

	if f.DoingSelection && !f.SelectionMoving {
		// Allow CommitSelection to detect a change
		f.MoveSelection(0, 0)
		f.CommitSelection()
	}

	cl.Redraw()
	f.RedrawRenderLayer()
}

// FlipHorizontal flips the layer horizontally, or flips the selection if anything
// is selected
func (f *File) FlipHorizontal() {
	latestHistory := HistoryPixel{make(map[IntVec2]PixelStateData), CurrentFile.CurrentLayer}

	var sx, sy int32 = 0, 0
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

	if f.DoingSelection && !f.SelectionMoving {
		// Allow CommitSelection to detect a change
		f.MoveSelection(0, 0)
	}

	cl.Redraw()
	f.RedrawRenderLayer()
}

// FlipVertical flips the layer vertically, or flips the selection if anything
// is selected
func (f *File) FlipVertical() {
	latestHistory := HistoryPixel{make(map[IntVec2]PixelStateData), CurrentFile.CurrentLayer}

	var sx, sy int32 = 0, 0
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

	if f.DoingSelection && !f.SelectionMoving {
		// Allow CommitSelection to detect a change
		f.MoveSelection(0, 0)
	}

	cl.Redraw()
	f.RedrawRenderLayer()
}

// Undo undoes an action
func (f *File) Undo() {
	if f.historyOffset < int32(len(f.History)) {
		f.historyOffset++
		index := int32(len(f.History)) - f.historyOffset
		history := f.History[index]

		var process func(historyItem interface{})
		process = func(historyItem interface{}) {
			switch typed := historyItem.(type) {
			case CompoundHistory:
				for i := 0; i < len(typed.Actions); i++ {
					process(typed.Actions[i])
				}
			case HistoryPixel:
				if f.DoingSelection {
					f.Selection = make(map[IntVec2]rl.Color)
					f.DoingSelection = false
					f.SelectionMoving = false
				}
				current := f.CurrentLayer
				f.SetCurrentLayer(typed.LayerIndex)
				layer := f.GetCurrentLayer()
				for pos, psd := range typed.PixelState {
					layer.PixelData[pos] = psd.Prev
				}
				layer.Redraw()
				f.SetCurrentLayer(current)
			case HistoryLayer:
				switch typed.HistoryLayerAction {
				case HistoryLayerActionDelete:
					f.RestoreLayer(typed.LayerIndex)
				case HistoryLayerActionCreate:
					f.DeleteLayer(typed.LayerIndex, false)
				case HistoryLayerActionMoveUp:
					f.MoveLayerUp(typed.LayerIndex, false)
				case HistoryLayerActionMoveDown:
					f.MoveLayerDown(typed.LayerIndex, false)
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
		}

		process(history)

		LayersUIRebuildList()
		f.RedrawRenderLayer()
	}
}

// Redo redoes an action
func (f *File) Redo() {
	if f.historyOffset > 0 {
		index := int32(len(f.History)) - f.historyOffset
		f.historyOffset--
		history := f.History[index]

		var process func(historyItem interface{})
		process = func(historyItem interface{}) {
			switch typed := historyItem.(type) {
			case CompoundHistory:
				for i := len(typed.Actions) - 1; i >= 0; i-- {
					process(typed.Actions[i])
				}
			case HistoryPixel:
				current := f.CurrentLayer
				f.SetCurrentLayer(typed.LayerIndex)
				layer := f.GetCurrentLayer()
				for pos, psd := range typed.PixelState {
					layer.PixelData[pos] = psd.Current
				}
				layer.Redraw()
				f.SetCurrentLayer(current)
			case HistoryLayer:
				switch typed.HistoryLayerAction {
				case HistoryLayerActionDelete:
					f.DeleteLayer(typed.LayerIndex, false)
				case HistoryLayerActionCreate:
					f.RestoreLayer(typed.LayerIndex)
				case HistoryLayerActionMoveUp:
					f.MoveLayerUp(typed.LayerIndex, false)
				case HistoryLayerActionMoveDown:
					f.MoveLayerDown(typed.LayerIndex, false)
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
		}

		process(history)

		LayersUIRebuildList()
		f.RedrawRenderLayer()
	}
}

// Destroy unloads each layer's canvas
func (f *File) Destroy() {
	for _, layer := range f.Layers {
		rl.UnloadRenderTexture(layer.Canvas)
	}

	for i, file := range Files {
		if file == f {
			Files = append(Files[:i], Files[i+1:]...)
			return
		}
	}
}

// SaveAs saves the file differently depending on the extension
func (f *File) SaveAs(path string) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ext := filepath.Ext(path)
	switch ext {
	case ".png":
		// Create a colored image of the given width and height.
		img := image.NewNRGBA(image.Rect(0, 0, int(f.CanvasWidth), int(f.CanvasHeight)))

		for x := int32(0); x < f.CanvasWidth; x++ {
			for y := int32(0); y < f.CanvasHeight; y++ {
				col := rl.Blank
				loc := IntVec2{x, y}
				for _, layer := range f.Layers[:len(f.Layers)-1] {
					if !layer.Hidden {
						if layerColor, ok := layer.PixelData[loc]; ok {
							col = BlendWithOpacity(col, layerColor, layer.BlendMode)
						}
					}
				}

				img.Set(int(x), int(y), color.NRGBA{
					col.R,
					col.G,
					col.B,
					col.A,
				})
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

	case ".pix":
		enc := gob.NewEncoder(file)

		gob.Register(rl.Color{})
		gob.Register(IntVec2{})

		fSer := &FileSer{
			DrawGrid:     f.DrawGrid,
			CanvasWidth:  f.CanvasWidth,
			CanvasHeight: f.CanvasHeight,
			TileWidth:    f.TileWidth,
			TileHeight:   f.TileHeight,
			Layers:       make([]*LayerSer, len(f.Layers)),
			Animations:   make([]*AnimationSer, len(f.Animations)),
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
		for a := range f.Animations {
			fSer.Animations[a] = &AnimationSer{
				Name:       f.Animations[a].Name,
				FrameStart: f.Animations[a].FrameStart,
				FrameEnd:   f.Animations[a].FrameEnd,
				Timing:     f.Animations[a].Timing,
			}
		}

		if err := enc.Encode(fSer); err != nil {
			log.Println(err)
		}

	default:
		log.Printf("Can't save: extension \"%s\" not supported\n", ext)
		return
	}

	// Change name in the tab
	spl := strings.Split(path, "/")
	f.Filename = spl[len(spl)-1]
	f.PathDir = strings.Join(spl[:len(spl)-1], "/")
	f.FileDir = path
	log.Println(f.Filename, f.PathDir, f.FileDir)
	f.FileChanged = false
	EditorsUIRebuild()
}

// Open a file
func Open(openPath string) *File {
	var f *File

	fi, err := os.Stat(openPath)
	if err != nil {
		log.Println(err)
	}
	if fi.Mode().IsRegular() {

		switch filepath.Ext(openPath) {
		case ".pix":
			reader, err := os.Open(openPath)
			if err != nil {
				log.Fatal(err)
			}
			defer reader.Close()
			dec := gob.NewDecoder(reader)
			fileSer := &FileSer{}
			if err := dec.Decode(&fileSer); err != nil {
				log.Println(err)
			}

			f = NewFile(fileSer.CanvasWidth, fileSer.CanvasHeight, fileSer.TileWidth, fileSer.TileHeight)
			f.PathDir = path.Dir(openPath)
			f.FileDir = openPath
			f.DrawGrid = fileSer.DrawGrid

			f.Layers = make([]*Layer, len(fileSer.Layers))
			for i, layer := range fileSer.Layers {
				f.Layers[i] = &Layer{
					Name:      layer.Name,
					Hidden:    layer.Hidden,
					PixelData: layer.PixelData,
					Width:     layer.Width,
					Height:    layer.Height,
					Canvas:    rl.LoadRenderTexture(layer.Width, layer.Height),
				}
				f.Layers[i].Redraw()
			}
			f.RenderLayer = NewLayer(f.CanvasWidth, f.CanvasHeight, "render", rl.Blank, true)
			f.Animations = make([]*Animation, len(fileSer.Animations))
			for i, animation := range fileSer.Animations {
				f.Animations[i] = &Animation{
					Name:       animation.Name,
					FrameStart: animation.FrameStart,
					FrameEnd:   animation.FrameEnd,
					Timing:     animation.Timing,
				}
			}

			spl := strings.Split(openPath, "/")
			f.Filename = spl[len(spl)-1]

			CurrentFile = f

			AnimationsUIRebuildList()
			LayersUIRebuildList()

		case ".png":
			tex := rl.LoadTexture(openPath)
			pixelColors := rl.LoadImageColors(rl.LoadImageFromTexture(tex))

			f = NewFile(tex.Width, tex.Height, 8, 8)
			f.PathDir = path.Dir(openPath)
			f.FileDir = openPath

			editedLayer := NewLayer(f.CanvasWidth, f.CanvasHeight, "background", rl.Blank, false)

			rl.BeginTextureMode(editedLayer.Canvas)
			for y := int32(0); y < f.CanvasHeight; y++ {
				for x := int32(0); x < f.CanvasWidth; x++ {
					color := pixelColors[x+y*f.CanvasWidth]
					editedLayer.PixelData[IntVec2{x, y}] = color
				}
			}
			f.RenderLayer = NewLayer(f.CanvasWidth, f.CanvasHeight, "render", rl.Blank, true)
			rl.EndTextureMode()
			editedLayer.Redraw()

			f.Layers = []*Layer{
				editedLayer,
				NewLayer(f.CanvasWidth, f.CanvasHeight, "hidden", rl.Blank, true),
			}

			spl := strings.Split(openPath, "/")
			f.Filename = spl[len(spl)-1]
		}
	}

	CurrentFile = f
	f.RedrawRenderLayer()
	EditorsUIRebuild()

	return f
}
