package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

type LayersUI struct {
	Position      IntVec2
	file          *File
	name          string
	Width, Height int
	Buttons       []*Button
}

func NewLayersUI(position IntVec2, width, height int, file *File, name string) *LayersUI {
	l := &LayersUI{
		Position: position,
		file:     file,
		name:     name,
		Width:    width,
		Height:   height,
		Buttons:  make([]*Button, 0, 16),
	}

	for i, layer := range l.file.Layers {
		_ = layer
		l.Buttons = append(l.Buttons, NewButton(
			rl.NewRectangle(float32(l.Position.X), float32(l.Position.Y+i*(32+4)), float32(l.Width), float32(32)),
			"Layer",
			func() {
				log.Println("clicked")
			}))
	}

	return l
}

func (l *LayersUI) MouseUp() {

}
func (l *LayersUI) MouseDown() {

}
func (l *LayersUI) Update() {
	for _, button := range l.Buttons {
		button.Update()
	}
}
func (l *LayersUI) Draw() {
	rl.DrawRectangle(l.Position.X, l.Position.Y, l.Width, l.Height, rl.Gray)
	for _, button := range l.Buttons {
		button.Draw()
	}
}
