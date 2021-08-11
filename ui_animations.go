package main

import rl "github.com/lachee/raylib-goplus/raylib"

var (
	currentAnimationHoverable *Hoverable
	animationInteractables    = make(map[int]*Entity)

	animationList          *Entity
	animationListContainer *Entity
)

// AnimationsUIRebuildList rebuilds the list
func AnimationsUIRebuildList() {
	animationList.DestroyNested()
	animationList.Destroy()
	animationListContainer.RemoveChild(animationList)

	if res, err := scene.QueryID(animationListContainer.ID); err == nil {
		moveable := res.Components[animationListContainer.Scene.ComponentsMap["moveable"]].(*Moveable)
		bounds := moveable.Bounds
		AnimationsUIMakeList(bounds)
		animationListContainer.PushChild(animationList)
		animationListContainer.FlowChildren()
	}
}

// AnimationsUIMakeBox makes a box for an animatio
func AnimationsUIMakeBox(y int, animation *Animation) *Entity {
	var bounds rl.Rectangle
	if res, err := scene.QueryID(animationListContainer.ID); err == nil {
		moveable := res.Components[animationListContainer.Scene.ComponentsMap["moveable"]].(*Moveable)
		bounds = moveable.Bounds
	}

	// moveUp := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2), GetFile("./res/icons/arrow_up.png"), false,
	// 	func(entity *Entity, button rl.MouseButton) {
	// 		// button up
	// 		if err := CurrentFile.MoveLayerUp(y); err == nil {
	// 			if CurrentFile.CurrentLayer == y {
	// 				CurrentFile.SetCurrentLayer(y + 1)
	// 			}
	// 			AnimationsUIRebuildList()
	// 		}
	// 	}, nil)
	// moveDown := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2), GetFile("./res/icons/arrow_down.png"), false,
	// 	func(entity *Entity, button rl.MouseButton) {
	// 		// button up
	// 		if err := CurrentFile.MoveLayerDown(y); err == nil {
	// 			if CurrentFile.CurrentLayer == y {
	// 				CurrentFile.SetCurrentLayer(y - 1)
	// 			}
	// 			AnimationsUIRebuildList()
	// 		}
	// 	}, nil)
	frameSelect := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2), GetFile("./res/icons/frame_selector.png"), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			// if err := CurrentFile.DeleteAnimation(y); err == nil {
			// 	AnimationsUIRebuildList()
			// }
		}, nil)
	delete := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2), GetFile("./res/icons/cross.png"), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			if err := CurrentFile.DeleteAnimation(y); err == nil {
				AnimationsUIRebuildList()
			}
		}, nil)

	// Keep the buttons organized
	buttonBox := NewBox(rl.NewRectangle(0, 0, UIButtonHeight*1.5, UIButtonHeight),
		[]*Entity{
			// moveUp,
			// moveDown,
			frameSelect,
			delete,
		},
		FlowDirectionHorizontal)

	isCurrent := CurrentFile.CurrentAnimation == y
	label := NewInput(rl.NewRectangle(0, 0, bounds.Width-UIButtonHeight*2.5, UIButtonHeight), animation.Name, isCurrent,
		func(entity *Entity, button rl.MouseButton) {
			// button up
		},
		func(entity *Entity, button rl.MouseButton, isHeld bool) {
			if entity == nil {
				// TODO find why the first call is nil
				return
			}
			if hoverable, ok := entity.GetHoverable(); ok {
				if currentAnimationHoverable != nil {
					currentAnimationHoverable.Selected = false
				}
				currentAnimationHoverable = hoverable
				hoverable.Selected = true

				CurrentFile.SetCurrentAnimation(y)
			}
		},
		func(entity *Entity, key rl.Key) {
			// key pressed
			if drawable, ok := entity.GetDrawable(); ok {
				if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
					if key == rl.KeyBackspace && len(drawableText.Label) > 0 {
						drawableText.Label = drawableText.Label[:len(drawableText.Label)-1]
					} else if len(drawableText.Label) < 12 {
						switch {
						// 0 to 9
						case key >= 48 && key <= 57:
							fallthrough
						// a to z
						case key >= 97 && key <= 97+26:
							fallthrough
						case key >= rl.KeyA && key <= rl.KeyZ:
							drawableText.Label += string(rune(key))
						}
					}
					CurrentFile.Animations[y].Name = drawableText.Label
				}
			}

		})

	// Set current animation ref
	if res, err := scene.QueryID(label.ID); err == nil {
		hoverable := res.Components[label.Scene.ComponentsMap["hoverable"]].(*Hoverable)

		if isCurrent {
			currentAnimationHoverable = hoverable
		}

		animationInteractables[y] = label
	}

	box := NewBox(rl.NewRectangle(0, 0, bounds.Width, UIButtonHeight), []*Entity{
		buttonBox,
		label,
	}, FlowDirectionHorizontal)
	return box
}

// AnimationsUIMakeList make a new list of animations
func AnimationsUIMakeList(bounds rl.Rectangle) {
	animationList = NewScrollableList(rl.NewRectangle(0, UIButtonHeight, bounds.Width, bounds.Height-UIButtonHeight), []*Entity{}, FlowDirectionVerticalReversed)
	// All of the animations
	for i, animation := range CurrentFile.Animations {

		animationList.PushChild(AnimationsUIMakeBox(i, animation))
	}
	animationList.FlowChildren()
}

// NewAnimationsUI creates the UI representation of the CurrentFile's animations
func NewAnimationsUI(bounds rl.Rectangle) *Entity {
	// New animation button
	newAnimationButton := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight), GetFile("./res/icons/plus.png"), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up

			CurrentFile.AddNewAnimation()
			animationList.PushChild(AnimationsUIMakeBox(0, CurrentFile.Animations[0]))
			animationList.FlowChildren()
		}, nil)

	animationListContainer = NewBox(bounds, []*Entity{
		newAnimationButton,
	}, FlowDirectionVertical)

	AnimationsUIMakeList(bounds)
	animationListContainer.PushChild(animationList)
	animationListContainer.FlowChildren()

	return animationListContainer
}
