package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// UIFileSystem handles non-ui drawing, including drawing the layer canvases
// TODO rename because it has nothing to do with FileSystem, only drawing the
// current File
type UIFileSystem struct {
	BasicSystem

	// Used for relational mouse movement
	mouseX, mouseY, mouseLastX, mouseLastY int32

	// workaround for resizing after AddSystem call has been made
	hasDoneFirstFrameResize bool

	cursor rl.Vector2
}

// NewUIFileSystem returns a new UIFileSystem
func NewUIFileSystem() *UIFileSystem {
	s := &UIFileSystem{
		hasDoneFirstFrameResize: false,
	}

	// Screen edges, left and top aren't needed since 0 stays constant
	screenRight := NewBlock(rl.NewRectangle(
		float32(rl.GetScreenWidth()),
		0,
		0,
		0,
	))
	if res, ok := screenRight.GetResizeable(); ok {
		res.OnResize = func(entity *Entity) {
			if mov, ok := entity.GetMoveable(); ok {
				// Move bounds to right side of the screen
				mov.Bounds.X = float32(rl.GetScreenWidth())
			}
		}
	}
	screenBottom := NewBlock(rl.NewRectangle(
		float32(rl.GetScreenWidth()),
		0,
		0,
		0,
	))
	if res, ok := screenBottom.GetResizeable(); ok {
		res.OnResize = func(entity *Entity) {
			if mov, ok := entity.GetMoveable(); ok {
				// Move bounds to right side of the screen
				mov.Bounds.Y = float32(rl.GetScreenHeight())
			}
		}
	}

	// Top bar
	menu := NewMenuUI(rl.NewRectangle(
		0,
		0,
		float32(rl.GetScreenWidth()),
		UIFontSize*2))

	editors := NewEditorsUI(rl.NewRectangle(
		0,
		0,
		float32(rl.GetScreenWidth()),
		UIFontSize*2))
	editors.Snap([]SnapData{
		{menu, SideTop, SideBottom},
	})

	// Right panel
	var rgbWidth = UIButtonHeight * 5.5
	var paletteWidth = UIButtonHeight * 3

	rightPanel := NewBox(rl.NewRectangle(0, 0, rgbWidth+paletteWidth, float32(rl.GetScreenHeight())),
		[]*Entity{}, FlowDirectionNone)
	rightPanel.Snap([]SnapData{
		{screenRight, SideRight, SideLeft},
	})
	if drawable, ok := rightPanel.GetDrawable(); ok {
		// drawable.DrawBorder = true
		drawable.DrawBackground = true
	}
	if res, ok := rightPanel.GetResizeable(); ok {
		res.OnResize = func(entity *Entity) {
			if moveable, ok := entity.GetMoveable(); ok {
				moveable.Bounds.Height = float32(rl.GetScreenHeight())
			}
		}
	}
	if interactable, ok := rightPanel.GetInteractable(); ok {
		interactable.OnMouseDown = func(entity *Entity, button MouseButton, isHeld bool) {

		}
	}

	rgb := NewRGBUI(rl.NewRectangle(
		0,
		0,
		rgbWidth,
		rgbWidth+UIButtonHeight*1.5))
	rgb.Snap([]SnapData{
		{screenRight, SideRight, SideLeft},
	})
	if res, ok := rgb.GetResizeable(); ok {
		res.OnResize = func(entity *Entity) {
			SetUIColors(LeftColor)
		}
	}

	palette := NewPaletteUI(rl.NewRectangle(
		0,
		0,
		paletteWidth,
		(rgbWidth+UIButtonHeight*1.5)-paletteWidth/4))
	palette.Snap([]SnapData{
		{rgb, SideRight, SideLeft},
	})

	currentColor := NewCurrentColorUI(rl.NewRectangle(
		0,
		0,
		paletteWidth,
		paletteWidth/4))
	currentColor.Snap([]SnapData{
		{rgb, SideRight, SideLeft},
		{palette, SideTop, SideBottom},
	})

	tools := NewToolsUI(rl.NewRectangle(
		0,
		0,
		rgbWidth+UIButtonHeight*2,
		UIButtonHeight))
	tools.Snap([]SnapData{
		{currentColor, SideLeft, SideLeft},
		{currentColor, SideTop, SideBottom},
	})

	layers := NewLayersUI(rl.NewRectangle(
		0,
		0,
		rgbWidth+paletteWidth,
		paletteWidth*2))
	layers.Snap([]SnapData{
		{screenRight, SideRight, SideLeft},
		{tools, SideTop, SideBottom},
	})
	if res, ok := layers.GetResizeable(); ok {
		res.OnResize = func(entity *Entity) {
			// Resize container and inner list to fill the remaining height
			if mov, ok := entity.GetMoveable(); ok {
				mov.Bounds.Height = float32(rl.GetScreenHeight()) - mov.Bounds.Y

				// Doesn't propagate, manually resize inner
				if lmov, ok := layerList.GetMoveable(); ok {
					lmov.Bounds.Height = float32(rl.GetScreenHeight()) - mov.Bounds.Y - UIButtonHeight
					if ldraw, ok := layerList.GetDrawable(); ok {
						if lparentDrawable, ok := ldraw.DrawableType.(*DrawableParent); ok {
							// Make a new texture for rendering the items to
							lparentDrawable.Texture = rl.LoadRenderTexture(int32(lmov.Bounds.Width), int32(lmov.Bounds.Height))
						}
					}
				}
			}
		}
	}

	// Left panel
	leftPanel := NewBox(rl.NewRectangle(0, 0, rgbWidth, float32(rl.GetScreenHeight())),
		[]*Entity{}, FlowDirectionNone)
	leftPanel.Snap([]SnapData{
		{editors, SideTop, SideBottom},
	})
	if drawable, ok := leftPanel.GetDrawable(); ok {
		// drawable.DrawBorder = true
		drawable.DrawBackground = true
	}
	if res, ok := leftPanel.GetResizeable(); ok {
		res.OnResize = func(entity *Entity) {
			if moveable, ok := entity.GetMoveable(); ok {
				moveable.Bounds.Height = float32(rl.GetScreenHeight())
			}
		}
	}
	if interactable, ok := leftPanel.GetInteractable(); ok {
		interactable.OnMouseDown = func(entity *Entity, button MouseButton, isHeld bool) {

		}
	}

	preview := NewPreviewUI(rl.NewRectangle(
		0,
		0,
		rgbWidth,
		rgbWidth,
	))
	preview.Snap([]SnapData{
		{editors, SideTop, SideBottom},
	})

	animations := NewAnimationsUI(rl.NewRectangle(
		0,
		0,
		rgbWidth,
		rgbWidth,
	))
	animations.Snap([]SnapData{
		{preview, SideTop, SideBottom},
	})
	if res, ok := animations.GetResizeable(); ok {
		res.OnResize = func(entity *Entity) {
			// Resize container and inner list to fill the remaining height
			if mov, ok := entity.GetMoveable(); ok {
				mov.Bounds.Height = float32(rl.GetScreenHeight()) - mov.Bounds.Y

				// Doesn't propagate, manually resize inner
				if lmov, ok := animationsList.GetMoveable(); ok {
					lmov.Bounds.Height = float32(rl.GetScreenHeight()) - mov.Bounds.Y - UIButtonHeight
					if ldraw, ok := animationsList.GetDrawable(); ok {
						if lparentDrawable, ok := ldraw.DrawableType.(*DrawableParent); ok {
							// Make a new texture for rendering the items to
							lparentDrawable.Texture = rl.LoadRenderTexture(int32(lmov.Bounds.Width), int32(lmov.Bounds.Height))
						}
					}
				}
			}
		}
	}

	NewResizeUI()

	return s
}

// Draw draws everything from the file to the screen
// TODO move all of this to system_render
func (s *UIFileSystem) Draw() {
	// Draw temp layer
	rl.BeginTextureMode(CurrentFile.Layers[len(CurrentFile.Layers)-1].Canvas)
	// LeftTool draws last as it's more important
	if rl.IsMouseButtonDown(rl.MouseRightButton) {
		RightTool.DrawPreview(int32(s.cursor.X), int32(s.cursor.Y))

	} else {
		LeftTool.DrawPreview(int32(s.cursor.X), int32(s.cursor.Y))
	}

	rl.EndTextureMode()

	// Draw layers
	rl.BeginMode2D(CurrentFile.FileCamera)
	for _, layer := range CurrentFile.Layers {
		if !layer.Hidden {
			rl.DrawTextureRec(layer.Canvas.Texture,
				rl.NewRectangle(0, 0, float32(layer.Canvas.Texture.Width), -float32(layer.Canvas.Texture.Height)),
				rl.NewVector2(-float32(layer.Canvas.Texture.Width)/2, -float32(layer.Canvas.Texture.Height)/2),
				rl.White)
		}
	}

	// Grid drawing
	if CurrentFile.DrawGrid {
		for x := int32(0); x <= CurrentFile.CanvasWidth; x += CurrentFile.TileWidth {
			rl.DrawLine(
				-CurrentFile.CanvasWidth/2+x,
				-CurrentFile.CanvasHeight/2,
				-CurrentFile.CanvasWidth/2+x,
				CurrentFile.CanvasHeight/2,
				rl.White)
		}
		for y := int32(0); y <= CurrentFile.CanvasHeight; y += CurrentFile.TileHeight {
			rl.DrawLine(
				-CurrentFile.CanvasWidth/2,
				-CurrentFile.CanvasHeight/2+y,
				CurrentFile.CanvasWidth/2,
				-CurrentFile.CanvasHeight/2+y,
				rl.White)
		}
	} else {
		// Draw canvas outline
		rl.DrawLine(
			-CurrentFile.CanvasWidth/2,
			-CurrentFile.CanvasHeight/2,
			CurrentFile.CanvasWidth/2,
			-CurrentFile.CanvasHeight/2,
			rl.White,
		)
		rl.DrawLine(
			-CurrentFile.CanvasWidth/2,
			-CurrentFile.CanvasHeight/2,
			-CurrentFile.CanvasWidth/2,
			CurrentFile.CanvasHeight/2,
			rl.White,
		)
		rl.DrawLine(
			-CurrentFile.CanvasWidth/2,
			CurrentFile.CanvasHeight/2,
			CurrentFile.CanvasWidth/2,
			CurrentFile.CanvasHeight/2,
			rl.White,
		)
		rl.DrawLine(
			CurrentFile.CanvasWidth/2,
			-CurrentFile.CanvasHeight/2,
			CurrentFile.CanvasWidth/2,
			CurrentFile.CanvasHeight/2,
			rl.White,
		)

	}

	// Show outline for canvas resize preview
	if CurrentFile.DoingResize {
		var x, y float32
		w := float32(CurrentFile.CanvasWidthResizePreview)
		h := float32(CurrentFile.CanvasHeightResizePreview)

		// Move offset
		dw := (w - float32(CurrentFile.CanvasWidth)) / 2
		dh := (h - float32(CurrentFile.CanvasHeight)) / 2

		switch CurrentFile.CanvasDirectionResizePreview {
		case ResizeTL:
			x = -float32(CurrentFile.CanvasWidthResizePreview)/2 + dw
			y = -float32(CurrentFile.CanvasHeightResizePreview)/2 + dh
		case ResizeTC:
			x = -float32(CurrentFile.CanvasWidthResizePreview) / 2
			y = -float32(CurrentFile.CanvasHeightResizePreview)/2 + dh
		case ResizeTR:
			x = -float32(CurrentFile.CanvasWidthResizePreview)/2 - dw
			y = -float32(CurrentFile.CanvasHeightResizePreview)/2 + dh
		case ResizeCL:
			x = -float32(CurrentFile.CanvasWidthResizePreview)/2 + dw
			y = -float32(CurrentFile.CanvasHeightResizePreview) / 2
		case ResizeCC:
			x = -float32(CurrentFile.CanvasWidthResizePreview) / 2
			y = -float32(CurrentFile.CanvasHeightResizePreview) / 2
		case ResizeCR:
			x = -float32(CurrentFile.CanvasWidthResizePreview)/2 - dw
			y = -float32(CurrentFile.CanvasHeightResizePreview) / 2
		case ResizeBL:
			x = -float32(CurrentFile.CanvasWidthResizePreview)/2 + dw
			y = -float32(CurrentFile.CanvasHeightResizePreview)/2 - dh
		case ResizeBC:
			x = -float32(CurrentFile.CanvasWidthResizePreview) / 2
			y = -float32(CurrentFile.CanvasHeightResizePreview)/2 - dh
		case ResizeBR:
			x = -float32(CurrentFile.CanvasWidthResizePreview)/2 - dw
			y = -float32(CurrentFile.CanvasHeightResizePreview)/2 - dh
		}

		rl.DrawRectangleLinesEx(
			rl.NewRectangle(x, y, w, h),
			1,
			rl.White,
		)
	}
	rl.EndMode2D()

	rl.BeginMode2D(rl.Camera2D{Zoom: 1.0})
	if rl.IsMouseButtonDown(rl.MouseRightButton) {
		RightTool.DrawUI(CurrentFile.FileCamera)
	} else {
		LeftTool.DrawUI(CurrentFile.FileCamera)
	}
	rl.EndMode2D()
}

func recursiveResize(entity *Entity) {
	if res, ok := entity.GetResizeable(); ok {
		if len(res.SnappedTo) > 0 {
			for _, snapData := range res.SnappedTo {
				recursiveResize(snapData.Parent)

				if childMoveable, ok := entity.GetMoveable(); ok {
					if parentMoveable, ok := snapData.Parent.GetMoveable(); ok {
						switch snapData.SnapSideChild {
						case SideLeft:
							childMoveable.Bounds.X = parentMoveable.Bounds.X
							if snapData.SnapSideParent == SideRight {
								childMoveable.Bounds.X += parentMoveable.Bounds.Width
							}
						case SideRight:
							childMoveable.Bounds.X = parentMoveable.Bounds.X
							if snapData.SnapSideParent == SideLeft {
								childMoveable.Bounds.X -= childMoveable.Bounds.Width
							}
						case SideTop:
							childMoveable.Bounds.Y = parentMoveable.Bounds.Y
							if snapData.SnapSideParent == SideBottom {
								childMoveable.Bounds.Y += parentMoveable.Bounds.Height
							}
						case SideBottom:
							childMoveable.Bounds.Y = parentMoveable.Bounds.Y
							if snapData.SnapSideParent == SideTop {
								childMoveable.Bounds.Y -= childMoveable.Bounds.Height
							}
						}
					}
				}
			}
		}

		entity.FlowChildren()

		if res.OnResize != nil {
			res.OnResize(entity)
		}
	}
}

// Resize is called when a resize event happens
func (s *UIFileSystem) Resize() {
	CurrentFile.FileCamera.Offset.X = float32(rl.GetScreenWidth()) / 2
	CurrentFile.FileCamera.Offset.Y = float32(rl.GetScreenHeight()) / 2

	s.hasDoneFirstFrameResize = true

	for _, result := range s.Scene.QueryTag(s.Scene.Tags["resizeable"], s.Scene.Tags["moveable"]) {
		_ = result
		recursiveResize(result.Entity)
	}

}

// Update updates the system
func (s *UIFileSystem) Update(dt float32) {
	// Move target
	if rl.IsWindowResized() || s.hasDoneFirstFrameResize == false {
		s.Resize()
	}

	layer := CurrentFile.GetCurrentLayer()
	s.mouseX = rl.GetMouseX()
	s.mouseY = rl.GetMouseY()

	// Scroll towards the cursor's location
	if !UIHasControl {
		scrollAmount := rl.GetMouseWheelMove()
		if scrollAmount != 0 {
			// TODO scroll scalar in config (0.1)
			CurrentFile.FileCameraTarget.X += ((float32(s.mouseX) - float32(rl.GetScreenWidth())/2) / (CurrentFile.FileCamera.Zoom * 10)) * float32(scrollAmount)
			CurrentFile.FileCameraTarget.Y += ((float32(s.mouseY) - float32(rl.GetScreenHeight())/2) / (CurrentFile.FileCamera.Zoom * 10)) * float32(scrollAmount)
			CurrentFile.FileCamera.Target = CurrentFile.FileCameraTarget
			CurrentFile.FileCamera.Zoom += float32(scrollAmount) * 0.1 * CurrentFile.FileCamera.Zoom
		}
	}

	if rl.IsMouseButtonDown(rl.MouseMiddleButton) {
		CurrentFile.FileCameraTarget.X += float32(s.mouseLastX-s.mouseX) / CurrentFile.FileCamera.Zoom
		CurrentFile.FileCameraTarget.Y += float32(s.mouseLastY-s.mouseY) / CurrentFile.FileCamera.Zoom
	}
	s.mouseLastX = s.mouseX
	s.mouseLastY = s.mouseY
	CurrentFile.FileCamera.Target = CurrentFile.FileCameraTarget

	s.cursor = rl.GetScreenToWorld2D(rl.GetMousePosition(), CurrentFile.FileCamera)
	s.cursor = rl.Vector2Add(
		s.cursor,
		rl.NewVector2(float32(layer.Canvas.Texture.Width)/2, float32(layer.Canvas.Texture.Height)/2),
	)

	PreviewUIDrawTile(int32(s.cursor.X), int32(s.cursor.Y))

	FileHasControl = false
	if !UIHasControl {
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {

			FileHasControl = true
			// Fires once
			if CurrentFile.HasDoneMouseUpLeft {
				// Create new history action
				switch LeftTool.(type) {
				case *PickerTool:
					// ignore
				case *SelectorTool:
					// ignore
				default:
					CurrentFile.AppendHistory(HistoryPixel{make(map[IntVec2]PixelStateData), CurrentFile.CurrentLayer})
				}
			}
			CurrentFile.HasDoneMouseUpLeft = false

			// Repeated action
			LeftTool.MouseDown(int32(s.cursor.X), int32(s.cursor.Y), rl.MouseLeftButton)
		} else {
			// Always fires once
			if CurrentFile.HasDoneMouseUpLeft == false {
				CurrentFile.HasDoneMouseUpLeft = true
				LeftTool.MouseUp(int32(s.cursor.X), int32(s.cursor.Y), rl.MouseLeftButton)
			}
		}

		if rl.IsMouseButtonDown(rl.MouseRightButton) {
			FileHasControl = true
			if CurrentFile.HasDoneMouseUpRight {
				// Create new history action
				switch LeftTool.(type) {
				case *PickerTool:
					// ignore
				case *SelectorTool:
					// ignore
				default:
					CurrentFile.AppendHistory(HistoryPixel{make(map[IntVec2]PixelStateData), CurrentFile.CurrentLayer})
				}
			}
			CurrentFile.HasDoneMouseUpRight = false
			RightTool.MouseDown(int32(s.cursor.X), int32(s.cursor.Y), rl.MouseRightButton)
		} else {
			if CurrentFile.HasDoneMouseUpRight == false {
				CurrentFile.HasDoneMouseUpRight = true
				RightTool.MouseUp(int32(s.cursor.X), int32(s.cursor.Y), rl.MouseRightButton)
			}
		}
	}
}
