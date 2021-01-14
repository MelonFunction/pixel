package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

type LayersUI struct {
	Position      IntVec2
	file          *File
	name          string
	Width, Height int
	Bounds        rl.Rectangle

	Components                []UIComponent
	scrollbar                 *Scroll
	ButtonWidth, ButtonHeight int

	Texture rl.RenderTexture2D

	wasMouseButtonDown bool
}

func NewLayersUI(position IntVec2, width, height int, file *File, name string) *LayersUI {
	l := &LayersUI{
		Position:           position,
		file:               file,
		name:               name,
		Width:              width,
		Height:             height,
		Bounds:             rl.NewRectangle(float32(position.X), float32(position.Y), float32(width), float32(height)),
		Components:         make([]UIComponent, 0, 16),
		ButtonWidth:        width - 20 - 16,
		ButtonHeight:       20,
		Texture:            rl.LoadRenderTexture(width, height),
		wasMouseButtonDown: false,
	}
	l.generateUI()
	return l
}

func (l *LayersUI) generateUI() {
	// TODO persistent state for things like scrollbar.elementOffset
	l.Components = make([]UIComponent, 0, 16)
	l.Components = append(l.Components, NewButton(
		rl.NewRectangle(0, 0, float32(l.ButtonHeight), float32(l.ButtonHeight)),
		Icon("./res/icons/plus.png"),
		true,
		func() {
			l.file.AddNewLayer()
			l.generateUI()
		}))

	scrollElements := make([]UIComponent, 0, 16)

	// Draw layer order in reverse
	max := len(l.file.Layers)
	for i := 0; i < len(l.file.Layers)-1; i++ {
		j := max - i - 2 // TODO why is this -2?
		m := i           // gotta make a new var since i's value changes each iteration

		var icon string
		if !l.file.Layers[m].Hidden {
			icon = "./res/icons/plus.png"
		}
		box := NewBox(
			rl.NewRectangle(0, float32(j*l.ButtonHeight), float32(l.ButtonWidth)+float32(l.ButtonHeight), float32(l.ButtonHeight)),
			[]UIComponent{
				NewButton(
					rl.NewRectangle(0, float32(j*l.ButtonHeight), float32(l.ButtonHeight), float32(l.ButtonHeight)),
					Icon(icon),
					true,
					func() {
						l.file.Layers[m].Hidden = !l.file.Layers[m].Hidden
						l.generateUI()
					}),
				NewButton(
					rl.NewRectangle(float32(l.ButtonHeight), float32(j*l.ButtonHeight), float32(l.ButtonWidth), float32(l.ButtonHeight)),
					Label(l.file.Layers[m].Name),
					i == l.file.CurrentLayer,
					func() {
						l.file.SetCurrentLayer(m)
						l.generateUI()
					}),
			},
			AlignHorizontal,
		)
		scrollElements = append(scrollElements, box)
	}

	l.scrollbar = NewScroll(
		rl.NewRectangle(float32(l.Position.X+l.Width-16), float32(l.Position.Y), 16, float32(l.Height)),
		rl.NewRectangle(float32(l.Position.X), float32(l.Position.Y+l.ButtonHeight), float32(l.Width-16), float32(l.Height-l.ButtonHeight)),
		scrollElements,
		len(l.file.Layers)-1,
		float32(l.ButtonHeight),
	)
}

func (l *LayersUI) GetWasMouseButtonDown() bool {
	return l.wasMouseButtonDown
}

func (l *LayersUI) SetWasMouseButtonDown(isDown bool) {
	l.wasMouseButtonDown = isDown
}

func (l *LayersUI) MouseUp() {
	if l == UIElementWithControl {
		// UIHasControl = false
		UIElementWithControl = nil
		UIComponentWithControl = nil // unset child too
	}
}

func (l *LayersUI) MouseDown() {
	// Using update instead since we have to check for hover on each component
}

func (l *LayersUI) CheckCollisions(offset rl.Vector2) bool {
	for _, component := range l.Components {
		if component.CheckCollisions(l.Bounds.Position()) {
			UIElementWithControl = l
			UIComponentWithControl = component
			return true
		}
	}
	if l.scrollbar.CheckCollisions(l.Bounds.Position()) {
		UIElementWithControl = l
		return true
	}

	return false
}

func (l *LayersUI) Update() {
}

func (l *LayersUI) Draw() {
	rl.BeginTextureMode(l.Texture)
	rl.ClearBackground(rl.Transparent)

	for _, component := range l.Components {
		component.Draw()
	}

	rl.EndTextureMode()

	l.scrollbar.Draw()

	rl.DrawTextureRec(l.Texture.Texture,
		rl.NewRectangle(0, 0, float32(l.Texture.Texture.Width), -float32(l.Texture.Texture.Height)),
		rl.NewVector2(float32(l.Position.X), float32(l.Position.Y)),
		rl.White)
}

func (l *LayersUI) Destroy() {
	l.Texture.Unload()
	for _, component := range l.Components {
		component.Destroy()
	}
	l.scrollbar.Destroy()
}
