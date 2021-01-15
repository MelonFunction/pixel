package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

type UIRenderSystem struct {
	BasicSystem
}

func NewUIRenderSystem() *UIRenderSystem {
	return &UIRenderSystem{}
}

func (s *UIRenderSystem) draw(d *Drawable, m *Moveable, h *Hoverable, isDrawingChildren bool) {
	// Don't render children
	if d.IsChild && !isDrawingChildren {
		return
	}

	if h.Hovered {
		rl.DrawRectangleRec(m.Bounds, rl.Black)
		rl.DrawRectangleLinesEx(m.Bounds, 2, rl.White)
	} else {
		if h.Selected {
			// TODO colorscheme
			// Same as hover for now
			rl.DrawRectangleRec(m.Bounds, rl.Black)
			rl.DrawRectangleLinesEx(m.Bounds, 2, rl.White)
		} else {
			rl.DrawRectangleRec(m.Bounds, rl.Black)
			rl.DrawRectangleLinesEx(m.Bounds, 2, rl.Gray)
		}
	}

	switch t := d.DrawableType.(type) {
	case *DrawableParent:
		rl.BeginTextureMode(t.Texture)
		for _, child := range t.Children {
			s.draw(child.Drawable, child.Moveable, child.Hoverable, true)
		}
		rl.EndTextureMode()
		rl.DrawTextureRec(t.Texture.Texture,
			rl.NewRectangle(0, 0, float32(t.Texture.Texture.Width), -float32(t.Texture.Texture.Height)),
			rl.NewVector2(m.Bounds.X, m.Bounds.Y),
			rl.White)
	case *DrawableText:
		fo := rl.MeasureTextEx(*Font, t.Label, 16, 1)
		x := m.Bounds.X + m.Bounds.Width/2 - fo.X/2
		y := m.Bounds.Y + m.Bounds.Height/2 - fo.Y/2
		rl.DrawTextEx(*Font, t.Label, rl.Vector2{X: x, Y: y}, 16, 1, rl.White)
	case *DrawableTexture:
		x := m.Bounds.X + m.Bounds.Width/2 - float32(t.Texture.Width)/2
		y := m.Bounds.Y + m.Bounds.Height/2 - float32(t.Texture.Height)/2
		rl.DrawTexture(t.Texture, int(x), int(y), rl.White)
	default:
		panic("Drawable not supported")
	}
}

func (s *UIRenderSystem) Update(dt float32) {
	for _, result := range s.Scene.Query(s.Scene.Tags["drawable, hoverable, moveable"]) {
		moveable := result.Components[s.Scene.ComponentsMap["moveable"]].(*Moveable)
		drawable := result.Components[s.Scene.ComponentsMap["drawable"]].(*Drawable)
		hoverable := result.Components[s.Scene.ComponentsMap["hoverable"]].(*Hoverable)

		s.draw(drawable, moveable, hoverable, false)
	}
}

type UIControlSystem struct {
	BasicSystem
}

func NewUIControlSystem() *UIControlSystem {
	return &UIControlSystem{}
}

func (s *UIControlSystem) Update(dt float32) {
	for _, result := range s.Scene.Query(s.Scene.Tags["basicControl"]) {
		moveable := result.Components[s.Scene.ComponentsMap["moveable"]].(*Moveable)
		hoverable := result.Components[s.Scene.ComponentsMap["hoverable"]].(*Hoverable)
		clickable := result.Components[s.Scene.ComponentsMap["clickable"]].(*Clickable)

		hoverable.Hovered = false
		if moveable.Bounds.Contains(rl.GetMousePosition()) {
			hoverable.Hovered = true

			if rl.IsMouseButtonDown(rl.MouseLeftButton) {
				UIHasControl = true
				clickable.ButtonDown = true
				clickable.OnMouseDown(rl.MouseLeftButton)
			} else if clickable.ButtonDown && !rl.IsMouseButtonDown(rl.MouseLeftButton) {
				clickable.ButtonDown = false
				clickable.OnMouseUp(rl.MouseLeftButton)
			}
		}

	}

	for _, result := range s.Scene.Query(s.Scene.Tags["scrollable"]) {
		scrollable := result.Components[s.Scene.ComponentsMap["scrollable"]].(*Scrollable)

		scrollAmount := rl.GetMouseWheelMove()
		if scrollAmount != 0 {
			UIHasControl = true
			scrollable.OnScroll(scrollAmount)
		}
	}

}
