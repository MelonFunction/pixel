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

	Buttons                   []*Button
	ButtonWidth, ButtonHeight int
}

func NewLayersUI(position IntVec2, width, height int, file *File, name string) *LayersUI {
	l := &LayersUI{
		Position:     position,
		file:         file,
		name:         name,
		Width:        width,
		Height:       height,
		Buttons:      make([]*Button, 0, 16),
		ButtonWidth:  width,
		ButtonHeight: 20,
	}
	l.generateUI()
	return l
}

func (l *LayersUI) generateUI() {
	l.Buttons = make([]*Button, 0, 16)
	l.Buttons = append(l.Buttons, NewButton(
		rl.NewRectangle(float32(l.Position.X), float32(l.Position.Y), float32(l.ButtonHeight), float32(l.ButtonHeight)),
		Icon("./res/icons/plus.png"),
		func() {
			l.file.Layers = append(
				[]*Layer{NewLayer(l.file.CanvasWidth, l.file.CanvasHeight, true)},
				l.file.Layers...)
			l.generateUI()
		}))
	// Offset by 1 position
	for i := 1; i < len(l.file.Layers); i++ {
		l.Buttons = append(l.Buttons, NewButton(
			rl.NewRectangle(float32(l.Position.X), float32(l.Position.Y+i*(l.ButtonHeight)), float32(l.ButtonWidth), float32(l.ButtonHeight)),
			Label(l.file.Layers[i].Name),
			func() {
				log.Println("clicked")
			}))
	}
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
	// rl.DrawRectangle(l.Position.X, l.Position.Y, l.Width, l.Height, rl.Gray)
	for _, button := range l.Buttons {
		button.Draw()
	}
}
