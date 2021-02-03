package main

import (
	"time"

	rl "github.com/lachee/raylib-goplus/raylib"
)

// KeymapData stores the action name as the key and a 2d slice of the keys
type KeymapData map[string][][]rl.Key

// Keymap stores the command+actions in Map and the the ordered keys in Keys
type Keymap struct {
	Keys []string
	Data KeymapData
}

// Static vars for file
var (
	keysExemptFromRelease = []rl.Key{
		rl.KeyLeftControl,
		rl.KeyLeftShift,
		rl.KeyLeftAlt,
		rl.KeyRightControl,
		rl.KeyRightShift,
		rl.KeyRightAlt,
	}
)

// NewKeymap returns a new Keymap
// It also sorts the keys to avoid conflicts between bindings as ctrl+z will
// fire before ctrl+shift+z if it is called first. Longer similar bindings will
// be before shorter similar ones in the list
func NewKeymap(data KeymapData) Keymap {
	keys := make([]string, 0, 0)

	for name, outer := range data {

		var longestInner []rl.Key
		for _, inner := range outer {
			if len(inner) > len(longestInner) {
				longestInner = inner
			}
		}
		didInsert := false
		for i, k := range keys {
			for _, inner := range data[k] {
				if len(longestInner) > len(inner) && !didInsert {
					didInsert = true
					keys = append(keys[:i], append([]string{name}, keys[i:]...)...)
				}
			}
		}

		if !didInsert {
			keys = append(keys, name)
		}
	}

	return Keymap{
		Keys: keys,
		Data: data,
	}
}

type UIControlSystem struct {
	BasicSystem
	file *File

	keyRepeatTimer      float32
	KeyRepeat           time.Duration
	Keymap              Keymap
	keyMoveable         bool
	lastKey             []rl.Key
	keysDown            map[rl.Key]bool // current keys down, used for combinations
	keysAwaitingRelease map[rl.Key]bool // keys which need to be released before they can be used again
}

func NewUIControlSystem(file *File, keymap Keymap) *UIControlSystem {
	return &UIControlSystem{
		file:                file,
		KeyRepeat:           time.Second / 5,
		Keymap:              keymap,
		keysDown:            make(map[rl.Key]bool),
		keysAwaitingRelease: make(map[rl.Key]bool),
	}
}

func (s *UIControlSystem) process(component interface{}, isProcessingChildren bool) {
	var result *QueryResult
	var entity *Entity
	switch typed := component.(type) {
	case *QueryResult:
		result = typed
		entity = typed.Entity
	case *Entity:
		entity = typed
		if res, err := scene.QueryID(typed.ID); err == nil {
			result = res
		}
	}

	drawable := result.Components[s.Scene.ComponentsMap["drawable"]].(*Drawable)
	moveable := result.Components[s.Scene.ComponentsMap["moveable"]].(*Moveable)
	hoverable := result.Components[s.Scene.ComponentsMap["hoverable"]].(*Hoverable)
	interactable := result.Components[s.Scene.ComponentsMap["interactable"]].(*Interactable)
	var scrollable *Scrollable
	scrollableInterface, ok := result.Components[s.Scene.ComponentsMap["scrollable"]]
	if ok {
		scrollable = scrollableInterface.(*Scrollable)
	}
	// hoverable.Hovered = false

	// Don't render children until the texture mode is set by the parent
	if drawable.IsChild && !isProcessingChildren {
		return
	}

	if moveable.Bounds.Contains(rl.GetMousePosition().Subtract(moveable.Offset)) {
		hoverable.Hovered = true
		switch t := drawable.DrawableType.(type) {
		case *DrawableParent:
			for _, child := range t.Children {
				s.process(child, true)
			}
		}

		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			UIHasControl = true
			interactable.ButtonDown = true
			if interactable.OnMouseDown != nil {
				UICompontentCapturedInput = interactable
				UIEntityCapturedInput = entity
				UIHasControl = true
				interactable.OnMouseDown(entity, rl.MouseLeftButton)
			}
		} else {
			if interactable.ButtonDown {
				interactable.ButtonDown = false
				if interactable.OnMouseUp != nil {
					interactable.OnMouseUp(entity, rl.MouseLeftButton)
				}
				UICompontentCapturedInput = nil
				UIEntityCapturedInput = nil
				UIHasControl = false
			}
		}

		if scrollable != nil {
			scrollAmount := rl.GetMouseWheelMove()
			if scrollAmount != 0 {
				UIHasControl = true
				scrollable.ScrollOffset += scrollAmount
			}
		}
	}
}

func (s *UIControlSystem) Update(dt float32) {
	// Handle keyboard events
	for key := range s.keysAwaitingRelease {
		if !rl.IsKeyDown(key) {
			delete(s.keysAwaitingRelease, key)
		}
	}

	checkDown := func(keySlices [][]rl.Key) bool {
		for _, keySlice := range keySlices {
			// Reset for each combination for the binding
			allDown := true
			for _, key := range keySlice {
				isDown := rl.IsKeyDown(key)
				s.keysDown[key] = isDown
				needsRelease, ok := s.keysAwaitingRelease[key]
				if !isDown || (ok && needsRelease) {
					allDown = false
				}
			}
			if allDown {
				return allDown
			}
		}
		return false
	}
	setAwaitingRelease := func(keySlices [][]rl.Key) bool {
		for _, keySlice := range keySlices {
			for _, key := range keySlice {
				found := false
				for _, k := range keysExemptFromRelease {
					if k == key {
						found = true
					}
				}
				if !found {
					s.keysAwaitingRelease[key] = true
				}
			}
		}
		return true
	}

	// If checkDown is true, then execute setAwaitingRelease (return isn't important)
	for _, key := range s.Keymap.Keys {
		if checkDown(s.Keymap.Data[key]) {
			setAwaitingRelease(s.Keymap.Data[key])

			switch key {
			case "layerUp":
				s.file.CurrentLayer++
				if s.file.CurrentLayer > len(s.file.Layers)-2 {
					s.file.CurrentLayer = len(s.file.Layers) - 2
				}
				LayersUISetCurrentLayer(s.file.CurrentLayer)
			case "layerDown":
				s.file.CurrentLayer--
				if s.file.CurrentLayer < 0 {
					s.file.CurrentLayer = 0
				}
				LayersUISetCurrentLayer(s.file.CurrentLayer)
			case "save":
				s.file.Save()
			case "export":
				s.file.Export()
			case "undo":
				s.file.Undo()
			case "redo":
				s.file.Redo()
			}

			break
		}

	}

	s.keyRepeatTimer += rl.GetFrameTime() * 1000
	if s.keyRepeatTimer > float32(s.KeyRepeat.Milliseconds()) {
		s.keyRepeatTimer = 0
		s.keyMoveable = true
	}
	// Stack keys up so that if left is held, then right is held, then right
	// is released, the cursor would continue going left instead of staying
	// still
	checkDownAddStack := func(keySlices [][]rl.Key) {
		for _, keySlice := range keySlices {
			for _, key := range keySlice {
				if rl.IsKeyPressed(key) {
					s.keyMoveable = true
					s.lastKey = append(s.lastKey, key)
				}
			}
		}
	}
	checkDownAddStack(s.Keymap.Data["toolRight"])
	checkDownAddStack(s.Keymap.Data["toolLeft"])
	checkDownAddStack(s.Keymap.Data["toolDown"])
	checkDownAddStack(s.Keymap.Data["toolUp"])

	if len(s.lastKey) > 0 && rl.IsKeyDown(s.lastKey[len(s.lastKey)-1]) {
		last := s.lastKey[len(s.lastKey)-1]
		if s.keyMoveable {
			s.keyRepeatTimer = 0
			s.keyMoveable = false

			moveAmount := int(fileSystem.Camera.Zoom)
			x := rl.GetMouseX()
			y := rl.GetMouseY()

			matches := func(match rl.Key, keySlices [][]rl.Key) bool {
				for _, keySlice := range keySlices {
					for _, key := range keySlice {
						if key == match {
							return true
						}
					}
				}
				return false
			}
			switch {
			case matches(last, s.Keymap.Data["toolRight"]):
				rl.SetMousePosition(x+moveAmount, y)
			case matches(last, s.Keymap.Data["toolLeft"]):
				rl.SetMousePosition(x-moveAmount, y)
			case matches(last, s.Keymap.Data["toolDown"]):
				rl.SetMousePosition(x, y+moveAmount)
			case matches(last, s.Keymap.Data["toolUp"]):
				rl.SetMousePosition(x, y-moveAmount)
			}
		}
	} else {
		// Pop lastKey until we find a key that's still down
		if len(s.lastKey) > 0 {
			s.lastKey = s.lastKey[:len(s.lastKey)-1]
		}
		s.keyRepeatTimer = 0
		s.keyMoveable = true
	}

	// Handle mouse events
	if rl.IsMouseButtonDown(rl.MouseLeftButton) && UICompontentCapturedInput != nil {
		UIHasControl = true
		if UICompontentCapturedInput != nil {
			if UICompontentCapturedInput.OnMouseDown != nil {
				UICompontentCapturedInput.OnMouseDown(UIEntityCapturedInput, MouseButtonNone)
			}
		}
	} else {
		UICompontentCapturedInput = nil
		UIEntityCapturedInput = nil
		UIHasControl = false

		for _, result := range s.Scene.QueryTag(s.Scene.Tags["basicControl"], s.Scene.Tags["scrollable"]) {
			s.process(result, false)
		}

	}
}
