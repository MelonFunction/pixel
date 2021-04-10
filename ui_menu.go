package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	// the buttons themselves
	menuButtons *Entity
	// the dropdown menu
	MenuContexts *Entity
)

type MenuContextOptions struct {
	Entity  *Entity
	Visible bool
}

func NewMenuUI(bounds rl.Rectangle) *Entity {
	menuButtons = NewScrollableList(bounds, []*Entity{}, FlowDirectionHorizontal)

	fo := rl.MeasureTextEx(*Font, "file", UIFontSize, 1)
	fileButton := NewButtonText(
		rl.NewRectangle(0, 0, fo.X+10, UIFontSize*2),
		"file", false, func(entity *Entity, button rl.MouseButton) {

		}, nil)
	menuButtons.PushChild(fileButton)
	if result, err := fileButton.Scene.QueryID(fileButton.ID); err == nil {
		hoverable, ok := result.Components[scene.ComponentsMap["hoverable"]].(*Hoverable)
		if ok {
			hoverable.OnMouseEnter = func() {
				log.Println("hovering")
			}
			hoverable.OnMouseLeave = func() {
				log.Println("not hovering")
			}
		}
	}

	fo = rl.MeasureTextEx(*Font, "save", UIFontSize, 1)
	saveButton := NewButtonText(
		rl.NewRectangle(0, UIFontSize*2, fo.X+10, UIFontSize*2),
		"save", false, func(entity *Entity, button rl.MouseButton) {
		}, nil)
	saveButton.Hide()
	fo = rl.MeasureTextEx(*Font, "open", UIFontSize, 1)
	openButton := NewButtonText(
		rl.NewRectangle(0, UIFontSize*4, fo.X+10, UIFontSize*2),
		"open", false, func(entity *Entity, button rl.MouseButton) {
		}, nil)
	openButton.Hide()
	fo = rl.MeasureTextEx(*Font, "resize", UIFontSize, 1)
	resizeButton := NewButtonText(
		rl.NewRectangle(0, UIFontSize*6, fo.X+10, UIFontSize*2),
		"resize", false, func(entity *Entity, button rl.MouseButton) {
		}, nil)
	resizeButton.Hide()

	MenuContexts = NewBox(bounds, []*Entity{
		saveButton,
		openButton,
		resizeButton,
	}, FlowDirectionNone)

	menuButtons.FlowChildren()
	return menuButtons
}
