package main

import (
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/gotk3/gotk3/gtk"
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

	UIControlSystemCmds    chan string
	UIControlSystemReturns chan string
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

	keyRepeatTimer      float32
	KeyRepeat           time.Duration
	Keymap              Keymap
	keyMoveable         bool
	lastKey             []rl.Key
	keysDown            map[rl.Key]bool // current keys down, used for combinations
	keysAwaitingRelease map[rl.Key]bool // keys which need to be released before they can be used again
}

func NewUIControlSystem(keymap Keymap) *UIControlSystem {
	UIControlSystemCmds = make(chan string)
	UIControlSystemReturns = make(chan string)
	go func(cmds, returns chan string) {
		gtk.Init(nil)

		win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
		if err != nil {
			log.Fatal("Unable to create window:", err)
		}
		win.Connect("destroy", func() {
			gtk.MainQuit()
			log.Println("destoryed")
		})

		// Only show png files
		filter, err := gtk.FileFilterNew()
		if err != nil {
			log.Fatal(err)
		}
		filter.AddPattern("*.png")

		// Default path is program exec location
		ex, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}

		running := true
		for running {
			select {
			case cmd := <-cmds:
				switch cmd {
				case "open":
					fc, err := gtk.FileChooserNativeDialogNew(
						"Select file to open",
						win,
						gtk.FILE_CHOOSER_ACTION_OPEN,
						"open",
						"cancel",
					)
					if err != nil {
						log.Fatal(err)
					}

					fc.AddFilter(filter)
					fc.SetCurrentFolder(filepath.Dir(ex))

					switch fc.Run() {
					case int(gtk.RESPONSE_ACCEPT):
						log.Println("accept")
						name := fc.GetFilename()
						log.Println(name)
						returns <- name
					default:
						log.Println("not accept")
						returns <- ""
					}
				case "save":
					fc, err := gtk.FileChooserNativeDialogNew(
						"Select file to save",
						win,
						gtk.FILE_CHOOSER_ACTION_SAVE,
						"save",
						"cancel",
					)
					if err != nil {
						log.Fatal(err)
					}

					fc.SetCurrentFolder(filepath.Dir(ex))

					switch fc.Run() {
					case int(gtk.RESPONSE_ACCEPT):
						log.Println("accept")
						name := fc.GetFilename()
						log.Println(name)
						returns <- name
					default:
						log.Println("not accept")
						returns <- ""
					}
				case "quit":
					running = false
					log.Println("quitting gtk")
					gtk.MainQuit()
				}
			default:
				time.Sleep(time.Millisecond * 100)
				gtk.MainIterationDo(false)
			}
		}
	}(UIControlSystemCmds, UIControlSystemReturns)

	return &UIControlSystem{
		KeyRepeat:           time.Second / 5,
		Keymap:              keymap,
		keysDown:            make(map[rl.Key]bool),
		keysAwaitingRelease: make(map[rl.Key]bool),
	}
}

func (s *UIControlSystem) getButtonDown() rl.MouseButton {
	button := MouseButtonNone
	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		button = rl.MouseLeftButton
	} else if rl.IsMouseButtonDown(rl.MouseRightButton) {
		button = rl.MouseRightButton
	} else if rl.IsMouseButtonDown(rl.MouseMiddleButton) {
		button = rl.MouseMiddleButton
	}

	return button
}

func (s *UIControlSystem) process(component interface{}, isProcessingChildren bool) bool {
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
	var interactable *Interactable
	interactableInterface, ok := result.Components[s.Scene.ComponentsMap["interactable"]]
	if ok {
		interactable = interactableInterface.(*Interactable)
	}
	var scrollable *Scrollable
	scrollableInterface, ok := result.Components[s.Scene.ComponentsMap["scrollable"]]
	if ok {
		scrollable = scrollableInterface.(*Scrollable)
	}
	// TODO see drawBorder in system_render.go
	// hoverable.Hovered = false

	// Don't render children until the texture mode is set by the parent
	if (drawable.IsChild && !isProcessingChildren) || drawable.Hidden {
		return false
	}

	if moveable.Bounds.Contains(rl.GetMousePosition().Subtract(moveable.Offset)) {
		hoverable.Hovered = true
		switch t := drawable.DrawableType.(type) {
		case *DrawableParent:
			for _, child := range t.Children {
				if s.process(child, true) {
					return true
				}
			}
		}

		// Scroll logic
		if scrollable != nil {
			scrollAmount := rl.GetMouseWheelMove()
			if scrollAmount != 0 {
				UIHasControl = true
				scrollable.ScrollOffset += scrollAmount
				return true
			}
		}

		// Mouse button logic
		button := s.getButtonDown()

		if interactable != nil && drawable.Hidden == false {
			if button != MouseButtonNone {
				// Mouse events are handled by the caller function (Update)
				if interactable.OnMouseDown != nil || interactable.OnMouseUp != nil {
					UIHasControl = true
					interactable.ButtonDown = button
					UIEntityCapturedInput = entity
					UIInteractableCapturedInput = interactable
					return true
				}
			}
		}
	}

	return false
}

func UIOpen() {
	UIControlSystemCmds <- "open"
	waiting := true
	for waiting {
		select {
		case name := <-UIControlSystemReturns:
			waiting = false
			if len(name) > 0 {
				file := Open(name)
				log.Println(file)
				Files = append(Files, file)
				CurrentFile = file
				EditorsUIAddButton(file)
			}
		}
	}
}

func UIExport() {
	UIControlSystemCmds <- "save"
	waiting := true
	for waiting {
		select {
		case name := <-UIControlSystemReturns:
			waiting = false
			if len(name) > 0 {
				CurrentFile.Export(name)
			}
		}
	}
}

func UISave() {
	UIControlSystemCmds <- "save"
	waiting := true
	for waiting {
		select {
		case name := <-UIControlSystemReturns:
			waiting = false
			if len(name) > 0 {
				CurrentFile.Save(name)
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

			// Don't allow events to happen if a component is inputting text
			if UIEntityCapturedInput != nil {
				break
			}

			switch key {
			case "toggleGrid":
				CurrentFile.DrawGrid = !CurrentFile.DrawGrid
			case "showDebug":
				ShowDebug = !ShowDebug
			case "resize":
				ResizeUIShowDialog()

			case "pixelBrush":
				// Simulate click event
				if interactable, ok := toolPencil.GetInteractable(); ok {
					interactable.OnMouseUp(toolPencil, rl.MouseRightButton)
				}
			case "eraser":
				if interactable, ok := toolEraser.GetInteractable(); ok {
					interactable.OnMouseUp(toolEraser, rl.MouseRightButton)
				}
			case "fill":
				if interactable, ok := toolFill.GetInteractable(); ok {
					interactable.OnMouseUp(toolFill, rl.MouseRightButton)
				}
			case "picker":
				if interactable, ok := toolPicker.GetInteractable(); ok {
					interactable.OnMouseUp(toolPicker, rl.MouseRightButton)
				}
			case "selector":
				if interactable, ok := toolSelector.GetInteractable(); ok {
					interactable.OnMouseUp(toolSelector, rl.MouseRightButton)
				}

			case "layerUp":
				CurrentFile.CurrentLayer++
				if CurrentFile.CurrentLayer > len(CurrentFile.Layers)-2 {
					CurrentFile.CurrentLayer = len(CurrentFile.Layers) - 2
				}
				LayersUISetCurrentLayer(CurrentFile.CurrentLayer)
			case "layerDown":
				CurrentFile.CurrentLayer--
				if CurrentFile.CurrentLayer < 0 {
					CurrentFile.CurrentLayer = 0
				}
				LayersUISetCurrentLayer(CurrentFile.CurrentLayer)
			case "open":
				UIOpen()
			case "save":
				UISave()
			case "export":
				UIExport()
			case "undo":
				CurrentFile.Undo()
			case "redo":
				CurrentFile.Redo()
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
				// Move selection
				if _, ok := CurrentFile.LeftTool.(*SelectorTool); ok {
					CurrentFile.MoveSelection(1, 0)
				} else {
					rl.SetMousePosition(x+moveAmount, y)
				}
			case matches(last, s.Keymap.Data["toolLeft"]):
				if _, ok := CurrentFile.LeftTool.(*SelectorTool); ok {
					CurrentFile.MoveSelection(-1, 0)
				} else {
					rl.SetMousePosition(x-moveAmount, y)
				}
			case matches(last, s.Keymap.Data["toolDown"]):
				if _, ok := CurrentFile.LeftTool.(*SelectorTool); ok {
					CurrentFile.MoveSelection(0, 1)
				} else {
					rl.SetMousePosition(x, y+moveAmount)
				}
			case matches(last, s.Keymap.Data["toolUp"]):
				if _, ok := CurrentFile.LeftTool.(*SelectorTool); ok {
					CurrentFile.MoveSelection(0, -1)
				} else {
					rl.SetMousePosition(x, y-moveAmount)
				}
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
	// shouldProcess prevents more events when a mouse button is down
	shouldProcess := true
	if UIEntityCapturedInput != nil {
		shouldProcess = false
		button := s.getButtonDown()
		lastButton := UIInteractableCapturedInput.ButtonDown

		if lastButton != MouseButtonNone && button == lastButton {
			// Continuously send mouse down event
			if UIInteractableCapturedInput.ButtonReleased == true {
				UIInteractableCapturedInput.ButtonDownAt = time.Now()
				UIInteractableCapturedInput.ButtonReleased = false
			}

			isHeld := false
			if time.Now().Sub(UIInteractableCapturedInput.ButtonDownAt) > time.Second/2 {
				isHeld = true
			}

			if UIInteractableCapturedInput.OnMouseDown != nil {
				UIInteractableCapturedInput.OnMouseDown(UIEntityCapturedInput, lastButton, isHeld)
			}
		} else if button == MouseButtonNone {
			// Handle mouse up event
			if UIInteractableCapturedInput.OnMouseUp != nil && UIInteractableCapturedInput.ButtonReleased == false {
				UIInteractableCapturedInput.ButtonReleased = true
				UIInteractableCapturedInput.OnMouseUp(UIEntityCapturedInput, lastButton)
			}

			if UIInteractableCapturedInput.OnKeyPress == nil {
				RemoveCapturedInput()
			}
		}

		// Handle keyboard events
		if UIInteractableCapturedInput != nil && UIInteractableCapturedInput.OnKeyPress != nil {
			shouldProcess = true
			lastKey := rl.GetKeyPressed()
			// GetKeyPressed doesn't get some keys for some reason
			if rl.IsKeyPressed(rl.KeyBackspace) {
				lastKey = rl.KeyBackspace
			}
			if rl.IsKeyPressed(rl.KeyEnter) {
				lastKey = rl.KeyEnter
			}
			if rl.IsKeyPressed(rl.KeyTab) {
				lastKey = rl.KeyTab
			}

			if uint32(lastKey) != math.MaxUint32 {
				UIInteractableCapturedInput.OnKeyPress(UIEntityCapturedInput, lastKey)
			}

			if button != MouseButtonNone {
				RemoveCapturedInput()
			}
		}

	}

	if shouldProcess {
		res := s.Scene.QueryTag(s.Scene.Tags["basic"], s.Scene.Tags["scrollable"], s.Scene.Tags["interactable"])
		// Reverse order so that entities that are on top can get input and return
		for i := len(res)/2 - 1; i >= 0; i-- {
			opp := len(res) - 1 - i
			res[i], res[opp] = res[opp], res[i]
		}

		for _, result := range res {
			// Clicked on something, stop processing
			if s.process(result, false) {
				return
			}
		}
		UIHasControl = false
	}

}
