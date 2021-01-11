package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	// UIHasControl lets the program know if input should go to the UI or not
	UIHasControl = false
)

type Button struct {
	bounds  rl.Rectangle
	onClick func()
	label   string
	hovered bool
}

func NewButton(bounds rl.Rectangle, label string, onClick func()) *Button {
	return &Button{
		bounds:  bounds,
		onClick: onClick,
		label:   label,
		hovered: false,
	}
}

func (b *Button) Update() {
	b.hovered = false

	if b.bounds.Contains(rl.GetMousePosition()) {
		b.hovered = true
		UIHasControl = true
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			b.onClick()
		}
	}
}

func (b *Button) Draw() {
	if b.hovered {
		rl.DrawRectangleRec(b.bounds, rl.Blue) // TODO get color from theme
	} else {
		rl.DrawRectangleRec(b.bounds, rl.White) // TODO get color from theme
	}
	rl.DrawText(b.label, int(b.bounds.X), int(b.bounds.Y), 16, rl.Black)
}
