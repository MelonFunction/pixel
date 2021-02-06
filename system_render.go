package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

type UIRenderSystem struct {
	BasicSystem
	camera rl.Camera2D
}

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

	log.Println("rendering", result.Entity.Name, result.Entity.ID)

	moveable := result.Components[s.Scene.ComponentsMap["moveable"]].(*Moveable)
	drawable := result.Components[s.Scene.ComponentsMap["drawable"]].(*Drawable)
	hoverable := result.Components[s.Scene.ComponentsMap["hoverable"]].(*Hoverable)
	var scrollable *Scrollable
	scrollableInterface, ok := result.Components[s.Scene.ComponentsMap["scrollable"]]
	if ok {
		scrollable = scrollableInterface.(*Scrollable)
	}

	// Don't render children until the texture mode is set by the parent
	if drawable.IsChild && !isDrawingChildren {
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
			rl.DrawRectangleRec(moveable.Bounds, rl.Black)
			rl.DrawRectangleLinesEx(moveable.Bounds, 2, rl.White)
		} else {
			if hoverable.Selected {
				// TODO colorscheme
				// Same as hover for now
				rl.DrawRectangleRec(moveable.Bounds, rl.Black)
				rl.DrawRectangleLinesEx(moveable.Bounds, 2, rl.White)
			} else {
				rl.DrawRectangleRec(moveable.Bounds, rl.Black)
				rl.DrawRectangleLinesEx(moveable.Bounds, 2, rl.Gray)
			}
		}
	}

	switch t := drawable.DrawableType.(type) {
	case *DrawableParent:
		if t.IsPassthrough {
			for _, child := range t.Children {
				// Just draw the child, offset is already set
				s.draw(child, true, offset)
			}
			return
		}

		rl.BeginTextureMode(t.Texture)
		rl.ClearBackground(rl.Transparent)

		// Offset all children the parent's position
		s.camera.Target.X = moveable.Bounds.X
		s.camera.Target.Y = moveable.Bounds.Y
		childOffset := rl.Vector2{}
		if scrollable != nil {
			switch scrollable.ScrollDirection {
			case ScrollDirectionVertical:
				s.camera.Target.Y -= float32(scrollable.ScrollOffset) * 16
				childOffset.Y = float32(scrollable.ScrollOffset) * 16
			case ScrollDirectionHorizontal:
				s.camera.Target.X -= float32(scrollable.ScrollOffset) * 16
				childOffset.X = float32(scrollable.ScrollOffset) * 16
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
		drawBorder(hoverable, moveable)
		fo := rl.MeasureTextEx(*Font, t.Label, UIFontSize, 1)
		x := moveable.Bounds.X + moveable.Bounds.Width/2 - fo.X/2
		y := moveable.Bounds.Y + moveable.Bounds.Height/2 - fo.Y/2
		rl.DrawTextEx(*Font, t.Label, rl.Vector2{X: x, Y: y}, UIFontSize, 1, rl.White)
	case *DrawableTexture:
		drawBorder(hoverable, moveable)
		x := moveable.Bounds.X + moveable.Bounds.Width/2 - float32(t.Texture.Width)/2
		y := moveable.Bounds.Y + moveable.Bounds.Height/2 - float32(t.Texture.Height)/2
		rl.DrawTexture(t.Texture, int(x), int(y), rl.White)
	case *DrawableRenderTexture:
		// drawBorder(hoverable, moveable)
		// maybe shrink texture to fit inside border instead of drawing on top?
		rl.DrawTextureRec(t.Texture.Texture,
			rl.NewRectangle(0, 0, float32(t.Texture.Texture.Width), -float32(t.Texture.Texture.Height)),
			rl.NewVector2(moveable.Bounds.X, moveable.Bounds.Y),
			rl.White)
	default:
		panic("Drawable not supported")
	}
}

func (s *UIRenderSystem) Draw() {
	for _, result := range s.Scene.QueryTag(s.Scene.Tags["drawable, hoverable, moveable"], s.Scene.Tags["scrollable"]) {
		s.draw(result, false, rl.Vector2{})
	}
}
