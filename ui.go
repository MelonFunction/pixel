package main

import (
	rl "github.com/lachee/raylib-goplus/raylib"
)

// UI is the interface for UI elements (they handle their own components + states)
type UI interface {
	CheckCollisions(offset rl.Vector2) bool // Offset is the parent UI position
	MouseDown()                             // Called each frame the mouse is down
	MouseUp()                               // Called once, when the mouse button is released
	GetWasMouseButtonDown() bool            // Ensures MouseUp is only called once
	SetWasMouseButtonDown(bool)

	Update()
	Draw()
	Destroy() // UI might use a texture for rendering to, destroy it before making a new one
}

type UIComponent interface {
	GetBounds() rl.Rectangle
	CheckCollisions(offset rl.Vector2) bool // Offset is the parent UI position
	Draw()
	Destroy()
}

var (
	// UIHasControl lets the program know if input should go to the UI or not
	UIHasControl = false
	// UIElementWithControl is the current element with control
	UIElementWithControl UI
	// UIComponentWithControl is the current ui component with control
	UIComponentWithControl UIComponent
	// isInited is a flag to record if InitUI has been called
	isInited = false
	// Font is the font used
	Font *rl.Font

	uiCamera               = rl.Camera2D{Zoom: 1}
	mouseX, mouseY         int
	mouseLastX, mouseLastY = -1, -1
)

// InitUI must be called before UI is used
func InitUI() {
	isInited = true
	Font = rl.LoadFont("./res/fonts/prstartk.ttf")
}

// AlignMode defines how elements should be aligned
type AlignMode int

const (
	// AlignHorizontal aligns horizontally
	AlignHorizontal AlignMode = iota
)

// Box can organise multiple elements within itself, depending on the AlignMode
type Box struct {
	bounds    rl.Rectangle
	elements  []UIComponent
	alignMode AlignMode
}

func NewBox(bounds rl.Rectangle, elements []UIComponent, alignMode AlignMode) *Box {
	b := &Box{
		bounds:    bounds,
		elements:  elements,
		alignMode: alignMode,
	}

	// TODO
	// Resize bounds depending on contents?
	// Reflow elements depending on alignMode
	switch alignMode {
	case AlignHorizontal:

	}

	return b
}
func (b *Box) GetBounds() rl.Rectangle {
	return b.bounds
}
func (b *Box) CheckCollisions(offset rl.Vector2) bool {
	for _, element := range b.elements {
		if element.CheckCollisions(offset) {
			return true
		}
	}
	return false
}
func (b *Box) Draw() {
	for _, element := range b.elements {
		element.Draw()
	}
}
func (b *Box) Destroy() {
	for _, element := range b.elements {
		element.Destroy()
	}
}

// Label is used for buttons with text labels
type Label string

// Icon is used for buttons with icon labels
type Icon string

// Button is a button UI element
type Button struct {
	bounds   rl.Rectangle
	onClick  func()
	hovered  bool
	selected bool // If white outline should be drawn

	isTextButton bool
	label        string
	icon         rl.Texture2D
}

func NewButton(bounds rl.Rectangle, label interface{}, selected bool, onClick func()) *Button {
	if !isInited {
		panic("Call InitUI")
	}
	b := &Button{
		bounds:   bounds,
		onClick:  onClick,
		hovered:  false,
		selected: selected,
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
func (b *Button) GetBounds() rl.Rectangle {
	return b.bounds
}

func (b *Button) CheckCollisions(offset rl.Vector2) bool {
	b.hovered = false

	if b.bounds.Contains(rl.GetMousePosition().Subtract(offset)) {
		b.hovered = true
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			b.onClick()
			return true
		}
	}
	return false
}

func (b *Button) Draw() {
	if b.hovered {
		rl.DrawRectangleRec(b.bounds, rl.Black)
		rl.DrawRectangleLinesEx(b.bounds, 2, rl.White)
	} else {
		if b.selected {
			// TODO colorscheme
			// Same as hover for now
			rl.DrawRectangleRec(b.bounds, rl.Black)
			rl.DrawRectangleLinesEx(b.bounds, 2, rl.White)
		} else {
			rl.DrawRectangleRec(b.bounds, rl.Black)
			rl.DrawRectangleLinesEx(b.bounds, 2, rl.Gray)
		}
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

func (b *Button) Destroy() {
	b.icon.Unload()
}

// Scroll is a scroll bar UI element
type Scroll struct {
	handleAreaBounds rl.Rectangle // Element movement area
	handleBounds     rl.Rectangle // Handle handleBounds
	elementBounds    rl.Rectangle // Where the scroll elements should be drawn
	topOffset        float32      // Acts like padding, like extra elements are in the elements slice

	elements      []UIComponent // All of the contained elements
	lines         int           // Could have multiple elements on the same row, so use a known value instead
	elementOffset int           // Offset by the dragged amount

	Texture rl.RenderTexture2D

	hovered bool
}

func NewScroll(handleAreaBounds, elementBounds rl.Rectangle, elements []UIComponent, lines int, topOffset float32) *Scroll {
	s := &Scroll{
		handleAreaBounds: handleAreaBounds,
		handleBounds:     handleAreaBounds,
		elementBounds:    elementBounds,
		elements:         elements,
		lines:            lines,
		topOffset:        topOffset,
		Texture:          rl.LoadRenderTexture(int(elementBounds.Width), int(elementBounds.Height)),
	}
	return s
}

func (s *Scroll) CheckCollisions(offset rl.Vector2) bool {
	s.hovered = false

	// UIComponentWithControl ownership feels a bit mangled
	// But maybe it's ok?

	offset = s.elementBounds.Position()
	offset.Y += float32(s.elementOffset)
	for _, component := range s.elements {
		if component.CheckCollisions(offset) {
			UIComponentWithControl = component
			return true
		}
	}

	// Doesn't need offset for some reason, TODO make it so it is consistent
	if s.handleAreaBounds.Contains(rl.GetMousePosition()) {
		s.hovered = true
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			UIComponentWithControl = s
			return true
		}
	}

	return false
}

func (s *Scroll) GetBounds() rl.Rectangle {
	return s.handleAreaBounds
}

func (s *Scroll) Draw() {
	rl.BeginTextureMode(s.Texture)
	rl.BeginMode2D(uiCamera)
	rl.ClearBackground(rl.Color{48, 48, 48, 255})

	elementHeight := s.topOffset
	for _, element := range s.elements {
		elementHeight += element.GetBounds().Height
	}

	if elementHeight > s.handleAreaBounds.Height {
		s.handleBounds.Height = s.handleAreaBounds.Height - (elementHeight - s.handleAreaBounds.Height)

		// Set minimum height to width
		// TODO config for this
		if s.handleBounds.Height < s.handleBounds.Width {
			s.handleBounds.Height = s.handleBounds.Width
		}

		// Offset
		mouseX, mouseY = rl.GetMouseX(), rl.GetMouseY()
		if UIComponentWithControl == s {
			if mouseLastY > -1 {
				s.handleBounds.Y -= float32(mouseLastY - mouseY)
			}
			if s.handleBounds.Y < s.handleAreaBounds.Y {
				s.handleBounds.Y += s.handleAreaBounds.Y - s.handleBounds.Y
			}
			if s.handleBounds.Y+s.handleBounds.Height > s.handleAreaBounds.Y+s.handleAreaBounds.Height {
				s.handleBounds.Y -= (s.handleBounds.Y + s.handleBounds.Height) - (s.handleAreaBounds.Y + s.handleAreaBounds.Height)
			}
			s.elementOffset = int(s.handleAreaBounds.Y - s.handleBounds.Y)
		}

		mouseLastX, mouseLastY = mouseX, mouseY
	}

	target := rl.Vector2{}
	target.Y -= float32(s.elementOffset)
	uiCamera.Target = target

	for _, element := range s.elements {
		element.Draw()
	}

	rl.EndMode2D()
	rl.EndTextureMode()

	rl.DrawTextureRec(s.Texture.Texture,
		rl.NewRectangle(0, 0, float32(s.Texture.Texture.Width), -float32(s.Texture.Texture.Height)),
		rl.NewVector2(float32(s.elementBounds.X), float32(s.elementBounds.Y)),
		rl.White)

	rl.DrawRectangleRec(s.handleBounds, rl.Gray)             // handle
	rl.DrawRectangleLinesEx(s.handleAreaBounds, 2, rl.White) // outline
}

func (s *Scroll) Destroy() {
	s.Texture.Unload()
}
