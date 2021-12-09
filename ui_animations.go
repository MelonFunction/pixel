package main

import (
	"log"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	currentAnimationHoverable *Hoverable
	animationInteractables    = make(map[int]*Entity)

	animationsList          *Entity
	animationsListContainer *Entity
)

// AnimationsUIRebuildList rebuilds the list
func AnimationsUIRebuildList() {
	animationsList.DestroyNested()
	animationsList.Destroy()
	animationsListContainer.RemoveChild(animationsList)

	if res, err := scene.QueryID(animationsListContainer.ID); err == nil {
		moveable := res.Components[animationsListContainer.Scene.ComponentsMap["moveable"]].(*Moveable)
		bounds := moveable.Bounds
		AnimationsUIMakeList(bounds)
		animationsListContainer.PushChild(animationsList)
		animationsListContainer.FlowChildren()
	} else {
		log.Println(err)
	}
}

// AnimationsUIMakeBox makes a box for an animatio
func AnimationsUIMakeBox(y int, animation *Animation) *Entity {
	var bounds rl.Rectangle
	if res, err := scene.QueryID(animationsListContainer.ID); err == nil {
		moveable := res.Components[animationsListContainer.Scene.ComponentsMap["moveable"]].(*Moveable)
		bounds = moveable.Bounds
	}

	frameSelect := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2), GetFile("./res/icons/frame_selector.png"), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			lastTool := CurrentFile.LeftTool
			CurrentFile.LeftTool = NewSpriteSelectorTool("Sprite Selector L", func(firstSprite, lastSprite int) {
				CurrentFile.LeftTool = lastTool

				CurrentFile.SetAnimationFrames(y, firstSprite, lastSprite)
			})
		}, nil)
	delete := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight/2, UIButtonHeight/2), GetFile("./res/icons/cross.png"), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			if err := CurrentFile.DeleteAnimation(y); err == nil {
				AnimationsUIRebuildList()
			}
		}, nil)

	// Keep the buttons organized
	buttonBox := NewBox(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight),
		[]*Entity{
			frameSelect,
			delete,
		},
		FlowDirectionHorizontal)

	isCurrent := CurrentFile.CurrentAnimation == y
	label := NewInput(rl.NewRectangle(0, 0, bounds.Width-UIButtonHeight, UIButtonHeight), animation.Name, TextAlignCenter, isCurrent,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			// Convert back into fps
			anim, err := CurrentFile.GetAnimation(y)
			if err != nil {
				log.Println(err)
				return
			}
			CurrentFile.SetCurrentAnimation(y)
			PreviewUISetTiming(anim.Timing)
			previewAnimationFrame = anim.FrameStart
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
			}
		},
		func(entity *Entity, key rl.Key) {
			// key pressed
			if drawable, ok := entity.GetDrawable(); ok {
				if drawableText, ok := drawable.DrawableType.(*DrawableText); ok {
					// TODO this could probably be added to util since the same
					// code exists in multiple places
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
					CurrentFile.SetAnimationName(y, drawableText.Label)
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
	animationsList = NewScrollableList(rl.NewRectangle(0, UIButtonHeight, bounds.Width, bounds.Height-UIButtonHeight), []*Entity{}, FlowDirectionVerticalReversed)
	// All of the animations
	for i, animation := range CurrentFile.Animations {
		animationsList.PushChild(AnimationsUIMakeBox(i, animation))
	}
	animationsList.FlowChildren()
}

// NewAnimationsUI creates the UI representation of the CurrentFile's animations
func NewAnimationsUI(bounds rl.Rectangle) *Entity {
	// New animation button
	newAnimationButton := NewButtonTexture(rl.NewRectangle(0, 0, UIButtonHeight, UIButtonHeight), GetFile("./res/icons/plus.png"), false,
		func(entity *Entity, button rl.MouseButton) {
			// button up
			CurrentFile.AddNewAnimation()
			AnimationsUIRebuildList()
		}, nil)

	animationsListContainer = NewBox(bounds, []*Entity{
		newAnimationButton,
	}, FlowDirectionVertical)

	AnimationsUIMakeList(bounds)
	animationsListContainer.PushChild(animationsList)
	animationsListContainer.FlowChildren()

	return animationsListContainer
}
