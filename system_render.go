package main

import (
	"fmt"

	rl "github.com/lachee/raylib-goplus/raylib"
)

// UIRenderSystem renders everything in the ECS
type UIRenderSystem struct {
	BasicSystem
	camera rl.Camera2D
}

// NewUIRenderSystem returs a new UIRenderSystem
func NewUIRenderSystem() *UIRenderSystem {
	return &UIRenderSystem{camera: rl.Camera2D{Zoom: 1}}
}

func (s *UIRenderSystem) draw(component interface{}, isDrawingChildren bool, offset rl.Vector2) {
	var result *QueryResult
	switch typed := component.(type) {
	case *QueryResult:
		result = typed
	case *Entity:
		if res, err := scene.QueryID(typed.ID); err == nil {
			result = res
		}
	}

	moveable := result.Components[s.Scene.ComponentsMap["moveable"]].(*Moveable)
	drawable := result.Components[s.Scene.ComponentsMap["drawable"]].(*Drawable)
	hoverable := result.Components[s.Scene.ComponentsMap["hoverable"]].(*Hoverable)
	var interactable *Interactable
	interactableInterface, ok := result.Components[s.Scene.ComponentsMap["interactable"]]
	if ok {
		interactable = interactableInterface.(*Interactable)
	}
	var scrollable *Scrollable
	scrollableInterface, ok := result.Components[s.Scene.ComponentsMap["scrollable"]]
	if ok {
		scrollable = scrollableInterface.(*Scrollable)
	}

	// Don't render children until the texture mode is set by the parent
	// Also don't render hidden components
	if (drawable.IsChild && !isDrawingChildren) || drawable.Hidden {
		return
	}

	// Set the offset, doesn't matter if element is a child or not
	moveable.Offset = offset

	drawBorder := func(hoverable *Hoverable, moveable *Moveable) {
		if hoverable.Hovered {
			// TODO find out why this is set false here instead of in the control
			// area. Elements aren't hovered when they are added via a button but they can
			// still be clicked etc. Appears that only hover is broken
			hoverable.Hovered = false
			rl.DrawRectangleLinesEx(moveable.Bounds, 2, rl.White)
		} else {
			if hoverable.Selected {
				// TODO colorscheme
				// Same as hover for now
				rl.DrawRectangleLinesEx(moveable.Bounds, 2, rl.White)
			} else {
				rl.DrawRectangleLinesEx(moveable.Bounds, 2, rl.Gray)
			}
		}
	}

	drawBackground := func(hoverable *Hoverable, moveable *Moveable) {
		if hoverable.Hovered {
			hoverable.Hovered = false
			rl.DrawRectangleRec(moveable.Bounds, rl.NewColor(0, 0, 0, 255*0.8))
		} else {
			if hoverable.Selected {
				// TODO colorscheme
				// Same as hover for now
				rl.DrawRectangleRec(moveable.Bounds, rl.NewColor(0, 0, 0, 255*0.8))
			} else {
				rl.DrawRectangleRec(moveable.Bounds, rl.NewColor(0, 0, 0, 255*0.8))
			}
		}
	}

	// Do OnMouseEnter callback
	if hoverable.Hovered && hoverable.DidMouseLeave {
		hoverable.DidMouseLeave = false
		if hoverable.OnMouseEnter != nil {
			hoverable.OnMouseEnter(result.Entity)
		}
	} else if !hoverable.Hovered && !hoverable.DidMouseLeave {
		hoverable.DidMouseLeave = true
		if hoverable.OnMouseLeave != nil {
			hoverable.OnMouseLeave(result.Entity)
		}
	}

	// debug bounds
	if ShowDebug {
		drawBorder(hoverable, moveable)
	}

	switch t := drawable.DrawableType.(type) {
	case *DrawableParent:
		if drawable.DrawBackground {
			drawBackground(hoverable, moveable)
		}
		if drawable.DrawBorder {
			drawBorder(hoverable, moveable)
		}

		if t.IsPassthrough {
			for _, child := range t.Children {
				// Just draw the child, offset is already set
				s.draw(child, true, offset)
			}
			return
		}

		rl.BeginTextureMode(t.Texture)
		rl.ClearBackground(rl.Transparent)

		// Offset all children by the parent's position
		s.camera.Target.X = moveable.Bounds.X
		s.camera.Target.Y = moveable.Bounds.Y
		childOffset := rl.Vector2{}
		if scrollable != nil {
			// TODO alter child offset positions here?
			switch scrollable.ScrollDirection {
			case ScrollDirectionVertical:
				s.camera.Target.Y -= float32(scrollable.ScrollOffset)
				childOffset.Y = float32(scrollable.ScrollOffset)
			case ScrollDirectionHorizontal:
				s.camera.Target.X -= float32(scrollable.ScrollOffset)
				childOffset.X = float32(scrollable.ScrollOffset)
			}
		}

		for _, child := range t.Children {
			rl.BeginMode2D(s.camera)
			s.draw(child, true, childOffset)
			rl.EndMode2D()
		}

		rl.EndTextureMode()

		rl.DrawTextureRec(t.Texture.Texture,
			rl.NewRectangle(0, 0, float32(t.Texture.Texture.Width), -float32(t.Texture.Texture.Height)),
			rl.NewVector2(moveable.Bounds.X, moveable.Bounds.Y),
			rl.White)

	case *DrawableText:
		if drawable.DrawBackground {
			drawBackground(hoverable, moveable)
		}
		if drawable.DrawBorder {
			drawBorder(hoverable, moveable)
		}

		text := t.Label
		if interactable != nil && UIInteractableCapturedInput == interactable && interactable.OnKeyPress != nil {
			text += "|"
		}

		fo := rl.MeasureTextEx(*Font, text, UIFontSize, 1)
		space := rl.MeasureTextEx(*Font, " ", UIFontSize, 1)
		var x, y float32
		switch t.TextAlign {
		case TextAlignLeft:
			x = moveable.Bounds.X + space.X
			y = moveable.Bounds.Y + moveable.Bounds.Height/2 - fo.Y/2
		case TextAlignRight:
			x = moveable.Bounds.X + moveable.Bounds.Width - fo.X - space.X
			y = moveable.Bounds.Y + moveable.Bounds.Height/2 - fo.Y/2
		case TextAlignCenter:
			x = moveable.Bounds.X + moveable.Bounds.Width/2 - fo.X/2
			y = moveable.Bounds.Y + moveable.Bounds.Height/2 - fo.Y/2
		}
		rl.DrawTextEx(*Font, text, rl.Vector2{X: x, Y: y}, UIFontSize, 1, rl.White)

	case *DrawableTexture:
		if drawable.DrawBackground {
			drawBackground(hoverable, moveable)
		}
		if drawable.DrawBorder {
			drawBorder(hoverable, moveable)
		}

		x := moveable.Bounds.X + moveable.Bounds.Width/2 - float32(t.Texture.Width)/2
		y := moveable.Bounds.Y + moveable.Bounds.Height/2 - float32(t.Texture.Height)/2
		rl.DrawTexture(t.Texture, int(x), int(y), rl.White)
	case *DrawableRenderTexture:
		// drawBorder(hoverable, moveable)
		// maybe shrink texture to fit inside border instead of drawing on top?
		rl.DrawTexturePro(t.Texture.Texture,
			rl.NewRectangle(0, 0, float32(t.Texture.Texture.Width), -float32(t.Texture.Texture.Height)),
			rl.NewRectangle(moveable.Bounds.X, moveable.Bounds.Y, moveable.Bounds.Width, moveable.Bounds.Height),
			rl.NewVector2(0, 0),
			0,
			rl.White)
	default:
		panic("Drawable not supported")
	}
}

// Update updates the system
func (s *UIRenderSystem) Update(dt float32) {}

// Draw draws the system
func (s *UIRenderSystem) Draw() {
	results := s.Scene.QueryTag(s.Scene.Tags["basic"], s.Scene.Tags["interactable"], s.Scene.Tags["scrollable"])
	for _, result := range results {
		s.draw(result, false, rl.Vector2{})
	}

	// Debug text
	if ShowDebug {
		incr := 20
		start := 80 - incr
		incrY := func() int {
			start += 20
			return start
		}

		rl.DrawText(fmt.Sprintf("UIHasControl: %v", UIHasControl), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("FileHasControl: %v", FileHasControl), 0, incrY(), 20, rl.White)

		rl.DrawText(fmt.Sprintf("CanvasWidthResizePreview: %v", CurrentFile.CanvasWidthResizePreview), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("CanvasHeightResizePreview: %v", CurrentFile.CanvasHeightResizePreview), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("TileWidthResizePreview: %v", CurrentFile.TileWidthResizePreview), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("TileHeightResizePreview: %v", CurrentFile.TileHeightResizePreview), 0, incrY(), 20, rl.White)

		rl.DrawText(fmt.Sprintf("UIInteractableCapturedInput: %v", UIInteractableCapturedInput), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("UIInteractableCapturedInputLast: %v", UIInteractableCapturedInputLast), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("UIEntityCapturedInput: %v", UIEntityCapturedInput), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("Current layer: %d", CurrentFile.CurrentLayer), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("HistoryOffset: %d", CurrentFile.historyOffset), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("History Len: %d", len(CurrentFile.History)), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("Colors: Left: %d, Right: %d", CurrentFile.LeftColor, CurrentFile.RightColor), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("Selection Len: %d", len(CurrentFile.Selection)), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("DoingSelection: %t", CurrentFile.DoingSelection), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("SelectionMoving: %t", CurrentFile.SelectionMoving), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("SelectionResizing: %t", CurrentFile.SelectionResizing), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("IsSelectionPasted: %t", CurrentFile.IsSelectionPasted), 0, incrY(), 20, rl.White)
		rl.DrawText(fmt.Sprintf("SelectionBounds: %d, %d, %d, %d", CurrentFile.SelectionBounds[0], CurrentFile.SelectionBounds[1], CurrentFile.SelectionBounds[2], CurrentFile.SelectionBounds[3]), 0, incrY(), 20, rl.White)
		// for y, history := range CurrentFile.History {
		// 	str := fmt.Sprintf("Layer: %d, Diff: %d",
		// 		history.LayerIndex,
		// 		len(history.PixelState))
		// 	rl.DrawText(str, 20, 20*y+260, 20, rl.White)
		// }
	}
}
