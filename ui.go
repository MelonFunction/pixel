package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	// UIHasControl lets the program know if input should go to the UI or not
	UIHasControl = false
	isInited     = false
	Font         *rl.Font
)

func InitUI() {
	isInited = true
	Font = rl.LoadFont("./res/fonts/prstartk.ttf")
}

type Label string
type Icon string

type Button struct {
	bounds  rl.Rectangle
	onClick func()
	hovered bool

	isTextButton bool
	label        string
	icon         rl.Texture2D
}

func NewButton(bounds rl.Rectangle, label interface{}, onClick func()) *Button {
	if !isInited {
		panic("Call InitUI")
	}
	b := &Button{
		bounds:  bounds,
		onClick: onClick,
		hovered: false,
	}

	switch d := label.(type) {
	case Label:
		b.label = string(d)
		b.isTextButton = true
	case Icon:
		b.icon = rl.LoadTexture(string(d))
		b.isTextButton = false
	default:
		panic("Unsupported type passed to NewButton")
	}

	return b
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
		rl.DrawRectangleRec(b.bounds, rl.Black)
		rl.DrawRectangleLinesEx(b.bounds, 2, rl.White)
	} else {
		rl.DrawRectangleRec(b.bounds, rl.Black)
		rl.DrawRectangleLinesEx(b.bounds, 2, rl.Gray)

	}
	if b.isTextButton {
		fo := rl.MeasureTextEx(*Font, b.label, 16, 1)
		x := b.bounds.X + b.bounds.Width/2 - fo.X/2
		y := b.bounds.Y + b.bounds.Height/2 - fo.Y/2
		rl.DrawTextEx(*Font, b.label, rl.Vector2{X: x, Y: y}, 16, 1, rl.White)
	} else {
		x := b.bounds.X + b.bounds.Width/2 - float32(b.icon.Width)/2
		y := b.bounds.Y + b.bounds.Height/2 - float32(b.icon.Height)/2
		rl.DrawTexture(b.icon, int(x), int(y), rl.White)
	}
}
