package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

// UI is the interface for UI elements (they handle their own components + states)
type UI interface {
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
	scene                                                               *Scene
	moveable, resizeable, interactable, hoverable, drawable, scrollable *Component
	renderSystem                                                        *UIRenderSystem
	controlSystem                                                       *UIControlSystem
)

// Moveable gives a component a position and dimensions
type Moveable struct {
	// Bounds is the position and dimensions of the component
	Bounds rl.Rectangle
	// OrigBounds is used when repositioning the element (stops offset stacking)
	OrigBounds rl.Rectangle
	// Offset values from scrolling
	Offset rl.Vector2
	// FlowDirection is how the elements should be arranged
	FlowDirection FlowDirection
}

// Resizeable allows a component to be resized and stores some callbacks
type Resizeable struct {
}

// Interactable is for storing all callbacks which can be procced by user inputs
// The callbacks are optional
type Interactable struct {
	// ButtonDown keeps track of if a button is down
	ButtonDown bool

	// OnMouseDown fires every frame the mouse button is down on the element
	OnMouseDown func(entity *Entity, button rl.MouseButton)
	// OnMouseUp fires once when the mouse is released (doesn't fire if mouse
	// is released while not within the bounds! Draggable should be used for
	// this kind of event instead)
	OnMouseUp func(entity *Entity, button rl.MouseButton)

	// OnScroll is for mouse wheel actions
	OnScroll func(direction int)
}

// ScrollDirection states the scroll direction of the component
type ScrollDirection int

const (
	// ScrollDirectionVertical is for vertical scrolling
	ScrollDirectionVertical ScrollDirection = iota
	// ScrollDirectionHorizontal is for horizontal scrolling
	ScrollDirectionHorizontal
)

// FlowDirection states which direction the children elements should flow in
type FlowDirection int

const (
	// FlowDirectionVertical flows vertically
	FlowDirectionVertical FlowDirection = iota
	// FlowDirectionVerticalReversed flows vertically, in reverse order
	FlowDirectionVerticalReversed
	// FlowDirectionHorizontal flows horizontally
	FlowDirectionHorizontal
	// FlowDirectionHorizontalReversed flows horizontally, in reverse order
	FlowDirectionHorizontalReversed
	// FlowDirectionNone doesn't reflow elements
	FlowDirectionNone
)

// Scrollable allows an element to render its children elements with an offset
type Scrollable struct {
	// ScrollDirection states which way the content should scroll
	ScrollDirection ScrollDirection
	// ScrollOffset is how much the content should be offset
	ScrollOffset int

	// TODO stuff for rendering scrollbars differently
}

// Hoverable stores the hovered and seleceted states
type Hoverable struct {
	Hovered  bool
	Selected bool
}

// Drawable handles all drawing related information
type Drawable struct {
	// DrawableType can be DrawableText, DrawableTexture or DrawableParent
	DrawableType interface{}

	// IsChild prevents normal rendering and instead renders to its
	// DrawableParent Texture
	IsChild bool
}

// DrawableText draws text
type DrawableText struct {
	Label string
}

type textureCache map[string]rl.Texture2D

func (d *DrawableTexture) SetTexture(path string) {
	d.Texture = rl.LoadTexture(path)
}

// DrawableTexture draws a texture
type DrawableTexture struct {
	Texture rl.Texture2D
}

func NewDrawableTexture(texturePath string) *DrawableTexture {
	d := &DrawableTexture{
		Texture: rl.LoadTexture(texturePath),
	}
	return d
}

// DrawableParent draws its children to its texture if IsPassthrough is true
type DrawableParent struct {
	// If true, doesn't draw to the Texture
	IsPassthrough bool
	Texture       rl.RenderTexture2D

	Children []*Entity
}

// InitUI must be called before UI is used
func InitUI() {
	isInited = true
	Font = rl.LoadFont("./res/fonts/prstartk.ttf")

	scene = NewScene()

	moveable = scene.NewComponent("moveable")
	resizeable = scene.NewComponent("resizeable")
	interactable = scene.NewComponent("interactable")
	scrollable = scene.NewComponent("scrollable")
	hoverable = scene.NewComponent("hoverable")
	drawable = scene.NewComponent("drawable")

	drawable.SetDestructor(func(e *Entity, data interface{}) {
		d, ok := data.(*Drawable)
		if ok {
			switch t := d.DrawableType.(type) {
			case *DrawableParent:
				if !t.IsPassthrough {
					t.Texture.Unload()
				}
			case *DrawableTexture:
				t.Texture.Unload()
			}
		}
	})

	scene.BuildTag("moveable", moveable)
	scene.BuildTag("resizeable", resizeable)
	scene.BuildTag("interactable", interactable)
	scene.BuildTag("scrollable", scrollable)
	scene.BuildTag("hoverable", hoverable)
	scene.BuildTag("drawable", drawable)
	scene.BuildTag("drawable, hoverable, moveable", drawable, moveable, hoverable)
	scene.BuildTag("basicControl", drawable, moveable, hoverable, interactable)

	controlSystem = NewUIControlSystem()
	renderSystem = NewUIRenderSystem()

	scene.AddSystem(controlSystem)
	scene.AddSystem(renderSystem)
}

func DestroyUI() {
	scene.Destroy()
}

// UpdateUI updates the systems (excluding the RenderSystem)
func UpdateUI() {
	UIHasControl = false
	controlSystem.Update(rl.GetFrameTime())
}

// DrawUI draws the RenderSystem
func DrawUI() {
	renderSystem.Update(rl.GetFrameTime())
}

// PushChild adds a child to a drawables children list and sets the relative
// initial positions of the children
func (e *Entity) PushChild(child *Entity) (*Entity, error) {
	log.Println("adding", child.Name, "to", e.Name)

	var err error
	if result, err := scene.QueryID(e.ID); err == nil {
		parentDrawable := result.Components[scene.ComponentsMap["drawable"]].(*Drawable)
		parentMoveable := result.Components[scene.ComponentsMap["moveable"]].(*Moveable)

		if result, err := scene.QueryID(child.ID); err == nil {
			childDrawable := result.Components[scene.ComponentsMap["drawable"]].(*Drawable)

			childDrawable.IsChild = true

			switch typed := parentDrawable.DrawableType.(type) {

			case *DrawableParent:
				found := false
				for _, c := range typed.Children {
					if c == child {
						found = true
					}
				}
				if !found {
					if parentMoveable.FlowDirection == FlowDirectionHorizontalReversed || parentMoveable.FlowDirection == FlowDirectionVerticalReversed {
						log.Println("\treversed")
						typed.Children = append([]*Entity{child}, typed.Children...)
					} else {
						log.Println("\tnot reversed")
						typed.Children = append(typed.Children, child)
					}
				}
			default:
				panic("Entity doesn't support child elements (make sure to only add children to boxes or scrolls!)")
			}

			switch typed := childDrawable.DrawableType.(type) {

			case *DrawableParent:
				for _, passChild := range typed.Children {
					child.PushChild(passChild)
				}
			}
		}
	}
	return nil, err
}

func (e *Entity) FlowChildren() {
	if result, err := scene.QueryID(e.ID); err == nil {
		parentDrawable := result.Components[scene.ComponentsMap["drawable"]].(*Drawable)
		parentMoveable := result.Components[scene.ComponentsMap["moveable"]].(*Moveable)

		children := make([]*Entity, 0, 16)

		switch typed := parentDrawable.DrawableType.(type) {

		case *DrawableParent:
			children = typed.Children
		default:
			panic("Entity doesn't support flowing as it doesn't have child elements (must be a box or scroll!)")
		}

		var fixNested func(e *Entity, parentDrawable *Drawable, parentMoveable *Moveable)
		fixNested = func(e *Entity, parentDrawable *Drawable, parentMoveable *Moveable) {
			children := make([]*Entity, 0, 16)

			switch typed := parentDrawable.DrawableType.(type) {

			case *DrawableParent:
				children = typed.Children
			default:
				return
			}

			for _, child := range children {
				if result, err := scene.QueryID(child.ID); err == nil {
					childDrawable := result.Components[scene.ComponentsMap["drawable"]].(*Drawable)
					childMoveable := result.Components[scene.ComponentsMap["moveable"]].(*Moveable)

					childMoveable.Bounds.X = parentMoveable.Bounds.X + childMoveable.OrigBounds.X
					childMoveable.Bounds.Y = parentMoveable.Bounds.Y + childMoveable.OrigBounds.Y

					fixNested(child, childDrawable, childMoveable)
				}
			}
		}

		_ = parentMoveable
		_ = children

		var offset rl.Vector2
		for _, child := range children {
			if result, err := scene.QueryID(child.ID); err == nil {
				childDrawable := result.Components[scene.ComponentsMap["drawable"]].(*Drawable)
				childMoveable := result.Components[scene.ComponentsMap["moveable"]].(*Moveable)

				childMoveable.Bounds.X = parentMoveable.Bounds.X
				childMoveable.Bounds.Y = parentMoveable.Bounds.Y

				if parentMoveable.FlowDirection == FlowDirectionVertical || parentMoveable.FlowDirection == FlowDirectionVerticalReversed {
					childMoveable.Bounds.Y += offset.Y
					offset.Y += childMoveable.Bounds.Height
				} else {
					childMoveable.Bounds.X += offset.X
					offset.X += childMoveable.Bounds.Width
				}

				fixNested(child, childDrawable, childMoveable)
			}
		}

	}
}

// NewButtonTexture creates a button which renders a texture
func NewButtonTexture(bounds rl.Rectangle, texturePath string, selected bool, onMouseUp, onMouseDown func(entity *Entity, button rl.MouseButton)) *Entity {
	e := scene.NewEntity(nil).
		AddComponent(moveable, &Moveable{bounds, bounds, rl.Vector2{}, FlowDirectionHorizontal}).
		AddComponent(hoverable, &Hoverable{Selected: selected}).
		AddComponent(interactable, &Interactable{OnMouseUp: onMouseUp, OnMouseDown: onMouseDown}).
		AddComponent(drawable, &Drawable{DrawableType: NewDrawableTexture(texturePath)})
	e.Name = "buttonTexture"
	return e
}

// NewButtonText creates a button which renders text
func NewButtonText(bounds rl.Rectangle, label string, selected bool, onMouseUp, onMouseDown func(entity *Entity, button rl.MouseButton)) *Entity {
	e := scene.NewEntity(nil).
		AddComponent(moveable, &Moveable{bounds, bounds, rl.Vector2{}, FlowDirectionHorizontal}).
		AddComponent(hoverable, &Hoverable{Selected: selected}).
		AddComponent(interactable, &Interactable{OnMouseUp: onMouseUp, OnMouseDown: onMouseDown}).
		AddComponent(drawable, &Drawable{DrawableType: &DrawableText{label}})
	e.Name = "buttonText: " + label
	return e
}

// prepareChildren moves children elements etc
func prepareChildren(entity *Entity, children []*Entity) {
	for _, child := range children {
		_, err := entity.PushChild(child)
		if err != nil {
			log.Println(err)
		}
	}
}

// NewBox creates a box which can store children
func NewBox(bounds rl.Rectangle, children []*Entity, flowDirection FlowDirection) *Entity {
	e := scene.NewEntity(nil).
		AddComponent(moveable, &Moveable{bounds, bounds, rl.Vector2{}, flowDirection}).
		AddComponent(hoverable, &Hoverable{Selected: false}).
		AddComponent(interactable, &Interactable{}).
		AddComponent(drawable, &Drawable{DrawableType: &DrawableParent{
			IsPassthrough: true,
			Children:      children,
		}})
	e.Name = "box"
	prepareChildren(e, children)
	e.FlowChildren()
	return e
}

// NewScrollableList creates a box, but it can scroll. Reversed is if the items
// order should be reversed
func NewScrollableList(bounds rl.Rectangle, children []*Entity, flowDirection FlowDirection) *Entity {
	e := scene.NewEntity(nil).
		AddComponent(moveable, &Moveable{bounds, bounds, rl.Vector2{}, flowDirection}).
		AddComponent(hoverable, &Hoverable{Selected: false}).
		AddComponent(interactable, &Interactable{}).
		AddComponent(scrollable, &Scrollable{}).
		AddComponent(drawable, &Drawable{DrawableType: &DrawableParent{
			IsPassthrough: false,
			Texture:       rl.LoadRenderTexture(int(bounds.Width), int(bounds.Height)),
		}})
	e.Name = "scroll"
	prepareChildren(e, children)
	e.FlowChildren()
	return e
}
