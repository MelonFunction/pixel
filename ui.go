package main

import (
	"log"

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

var (
	// UIHasControl lets the program know if input should go to the UI or not
	UIHasControl = false
	// UIElementWithControl is the current element with control
	UIElementWithControl UI
	// UIComponentWithControl is the current ui component with control
	// UIComponentWithControl UIComponent
	// isInited is a flag to record if InitUI has been called
	isInited = false
	// Font is the font used
	Font *rl.Font

	uiCamera               = rl.Camera2D{Zoom: 1}
	mouseX, mouseY         int
	mouseLastX, mouseLastY = -1, -1

	// Ecs stuffs
	scene                                                            *Scene
	moveable, resizeable, clickable, hoverable, drawable, scrollable *Component
	renderSystem                                                     *UIRenderSystem
	controlSystem                                                    *UIControlSystem
)

type Moveable struct {
	Bounds rl.Rectangle
}

type Resizeable struct {
}

type Clickable struct {
	ButtonDown bool

	OnMouseDown func(button rl.MouseButton)
	OnMouseUp   func(button rl.MouseButton)
}

type Scrollable struct {
	OnScroll func(dir int)
}

type Hoverable struct {
	Hovered  bool
	Selected bool
}

type Drawable struct {
	// DrawableType can be DrawableText, DrawableTexture or DrawableParent
	DrawableType interface{}

	// IsChild prevents normal rendering and instead renders to its
	// DrawableParent Texture
	IsChild bool
}

type DrawableText struct {
	Label string
}

type DrawableTexture struct {
	Texture rl.Texture2D
}

// DrawableChild is just a quick reference to the components the parent needs
// to draw it
type DrawableChild struct {
	Drawable  *Drawable
	Moveable  *Moveable
	Hoverable *Hoverable
}

type DrawableParent struct {
	Texture  rl.RenderTexture2D
	Children []DrawableChild
}

// InitUI must be called before UI is used
func InitUI() {
	isInited = true
	Font = rl.LoadFont("./res/fonts/prstartk.ttf")

	scene = NewScene()

	moveable = scene.NewComponent("moveable")
	resizeable = scene.NewComponent("resizeable")
	clickable = scene.NewComponent("clickable")
	scrollable = scene.NewComponent("scrollable")
	hoverable = scene.NewComponent("hoverable")
	drawable = scene.NewComponent("drawable")

	drawable.SetDestructor(func(e *Entity, data interface{}) {
		d, ok := data.(*Drawable)
		if ok {
			switch t := d.DrawableType.(type) {
			case *DrawableTexture:
				log.Println("unloading")
				t.Texture.Unload()
			}
		}
	})

	scene.BuildTag("moveable", moveable)
	scene.BuildTag("resizeable", resizeable)
	scene.BuildTag("clickable", clickable)
	scene.BuildTag("scrollable", scrollable)
	scene.BuildTag("hoverable", hoverable)
	scene.BuildTag("drawable", drawable)
	scene.BuildTag("drawable, hoverable, moveable", drawable, moveable, hoverable)
	scene.BuildTag("basicControl", clickable, hoverable, moveable)

	renderSystem = NewUIRenderSystem()
	controlSystem = NewUIControlSystem()

	scene.AddSystem(renderSystem)
	scene.AddSystem(controlSystem)
}

func UpdateUI() {
	UIHasControl = false
	controlSystem.Update(rl.GetFrameTime())
}

func DrawUI() {
	renderSystem.Update(rl.GetFrameTime())
}

func NewButtonTexture(bounds rl.Rectangle, texturePath string, selected bool, onMouseUp, onMouseDown func(button rl.MouseButton)) *Entity {
	texture := rl.LoadTexture(string(texturePath))
	return scene.NewEntity().
		AddComponent(moveable, &Moveable{bounds}).
		AddComponent(hoverable, &Hoverable{Selected: selected}).
		AddComponent(clickable, &Clickable{OnMouseUp: onMouseUp, OnMouseDown: onMouseDown}).
		AddComponent(drawable, &Drawable{DrawableType: &DrawableTexture{texture}})
}

func NewButtonText(bounds rl.Rectangle, label string, selected bool, onMouseUp, onMouseDown func(button rl.MouseButton)) *Entity {
	return scene.NewEntity().
		AddComponent(moveable, &Moveable{bounds}).
		AddComponent(hoverable, &Hoverable{Selected: selected}).
		AddComponent(clickable, &Clickable{OnMouseUp: onMouseUp, OnMouseDown: onMouseDown}).
		AddComponent(drawable, &Drawable{DrawableType: &DrawableText{label}})

}

func NewBox(bounds rl.Rectangle, children []*Entity) *Entity {
	drawables := make([]DrawableChild, 0, 8)
	for _, child := range children {
		for _, result := range scene.Query(child.ID) {
			drawable := result.Components[scene.ComponentsMap["drawable"]].(*Drawable)
			moveable := result.Components[scene.ComponentsMap["moveable"]].(*Moveable)
			hoverable := result.Components[scene.ComponentsMap["hoverable"]].(*Hoverable)

			drawable.IsChild = true
			drawables = append(drawables, DrawableChild{
				Drawable:  drawable,
				Moveable:  moveable,
				Hoverable: hoverable,
			})
		}
	}

	// Not actually going to make boxes DrawableParents, this is just testing

	return scene.NewEntity().
		AddComponent(moveable, &Moveable{bounds}).
		AddComponent(hoverable, &Hoverable{Selected: false}).
		AddComponent(drawable, &Drawable{DrawableType: &DrawableParent{
			Texture:  rl.LoadRenderTexture(int(bounds.Width), int(bounds.Height)),
			Children: drawables,
		}})
}

// // Box can organise multiple elements within itself, depending on the AlignMode
// type Box struct {
// 	bounds   rl.Rectangle
// 	elements []UIComponent
// }

// func NewBox(bounds rl.Rectangle, elements []UIComponent) *Box {
// 	b := &Box{
// 		bounds:   bounds,
// 		elements: elements,
// 	}

// 	return b
// }
// func (b *Box) GetBounds() rl.Rectangle {
// 	return b.bounds
// }
// func (b *Box) CheckCollisions(offset rl.Vector2) bool {
// 	for _, element := range b.elements {
// 		if element.CheckCollisions(offset) {
// 			return true
// 		}
// 	}
// 	return false
// }
// func (b *Box) Draw() {
// 	for _, element := range b.elements {
// 		element.Draw()
// 	}
// }
// func (b *Box) Destroy() {
// 	for _, element := range b.elements {
// 		element.Destroy()
// 	}
// }

// // Label is used for buttons with text labels
// type Label string

// // Icon is used for buttons with icon labels
// type Icon string

// // Scroll is a scroll bar UI element
// type Scroll struct {
// 	handleAreaBounds rl.Rectangle // Element movement area
// 	handleBounds     rl.Rectangle // Handle handleBounds
// 	elementBounds    rl.Rectangle // Where the scroll elements should be drawn
// 	topOffset        float32      // Acts like padding, like extra elements are in the elements slice

// 	elements      []UIComponent // All of the contained elements
// 	lines         int           // Could have multiple elements on the same row, so use a known value instead
// 	elementOffset int           // Offset by the dragged amount

// 	Texture rl.RenderTexture2D

// 	hovered bool
// }

// func NewScroll(handleAreaBounds, elementBounds rl.Rectangle, elements []UIComponent, lines int, topOffset float32) *Scroll {
// 	s := &Scroll{
// 		handleAreaBounds: handleAreaBounds,
// 		handleBounds:     handleAreaBounds,
// 		elementBounds:    elementBounds,
// 		elements:         elements,
// 		lines:            lines,
// 		topOffset:        topOffset,
// 		Texture:          rl.LoadRenderTexture(int(elementBounds.Width), int(elementBounds.Height)),
// 	}
// 	return s
// }

// func (s *Scroll) CheckCollisions(offset rl.Vector2) bool {
// 	s.hovered = false

// 	// UIComponentWithControl ownership feels a bit mangled
// 	// But maybe it's ok?

// 	offset = s.elementBounds.Position()
// 	offset.Y += float32(s.elementOffset)
// 	for _, component := range s.elements {
// 		if component.CheckCollisions(offset) {
// 			UIComponentWithControl = component
// 			return true
// 		}
// 	}

// 	// Doesn't need offset for some reason, TODO make it so it is consistent
// 	if s.handleAreaBounds.Contains(rl.GetMousePosition()) {
// 		s.hovered = true
// 		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
// 			UIComponentWithControl = s
// 			return true
// 		}
// 	}

// 	return false
// }

// func (s *Scroll) GetBounds() rl.Rectangle {
// 	return s.handleAreaBounds
// }

// func (s *Scroll) Draw() {
// 	rl.BeginTextureMode(s.Texture)
// 	rl.BeginMode2D(uiCamera)
// 	rl.ClearBackground(rl.Color{48, 48, 48, 255})

// 	elementHeight := s.topOffset
// 	for _, element := range s.elements {
// 		elementHeight += element.GetBounds().Height
// 	}

// 	if elementHeight > s.handleAreaBounds.Height {
// 		s.handleBounds.Height = s.handleAreaBounds.Height - (elementHeight - s.handleAreaBounds.Height)

// 		// Set minimum height to width
// 		// TODO config for this
// 		if s.handleBounds.Height < s.handleBounds.Width {
// 			s.handleBounds.Height = s.handleBounds.Width
// 		}

// 		// Offset
// 		mouseX, mouseY = rl.GetMouseX(), rl.GetMouseY()
// 		if UIComponentWithControl == s {
// 			if mouseLastY > -1 {
// 				s.handleBounds.Y -= float32(mouseLastY - mouseY)
// 			}
// 			if s.handleBounds.Y < s.handleAreaBounds.Y {
// 				s.handleBounds.Y += s.handleAreaBounds.Y - s.handleBounds.Y
// 			}
// 			if s.handleBounds.Y+s.handleBounds.Height > s.handleAreaBounds.Y+s.handleAreaBounds.Height {
// 				s.handleBounds.Y -= (s.handleBounds.Y + s.handleBounds.Height) - (s.handleAreaBounds.Y + s.handleAreaBounds.Height)
// 			}
// 			s.elementOffset = int(s.handleAreaBounds.Y - s.handleBounds.Y)
// 		}

// 		mouseLastX, mouseLastY = mouseX, mouseY
// 	}

// 	target := rl.Vector2{}
// 	target.Y -= float32(s.elementOffset)
// 	uiCamera.Target = target

// 	for _, element := range s.elements {
// 		element.Draw()
// 	}

// 	rl.EndMode2D()
// 	rl.EndTextureMode()

// 	rl.DrawTextureRec(s.Texture.Texture,
// 		rl.NewRectangle(0, 0, float32(s.Texture.Texture.Width), -float32(s.Texture.Texture.Height)),
// 		rl.NewVector2(float32(s.elementBounds.X), float32(s.elementBounds.Y)),
// 		rl.White)

// 	rl.DrawRectangleRec(s.handleBounds, rl.Gray)             // handle
// 	rl.DrawRectangleLinesEx(s.handleAreaBounds, 2, rl.White) // outline
// }

// func (s *Scroll) Destroy() {
// 	s.Texture.Unload()
// }
