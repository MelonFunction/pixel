package main

import (
	"log"
	"math"
	"time"

	"github.com/gotk3/gotk3/gtk"
	rl "github.com/lachee/raylib-goplus/raylib"
)

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

type UIControlSystem struct {
	BasicSystem

	keyRepeatTimer      float32
	KeyRepeat           time.Duration
	Keymap              Keymap
	keyMoveable         bool
	lastKey             []rl.Key
	mouseButtonDown     bool
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
		})

		// Only show png files
		filter, err := gtk.FileFilterNew()
		if err != nil {
			log.Fatal(err)
		}
		filter.SetName(".png, .pix")
		filter.AddPattern("*.png")
		filter.AddPattern("*.pix")

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
					fc.SetCurrentFolder(CurrentFile.PathDir)

					switch fc.Run() {
					case int(gtk.RESPONSE_ACCEPT):
						name := fc.GetFilename()
						log.Println(name)
						returns <- name
					default:
						returns <- ""
					}
					fc.Destroy()

				case "export":
					fc, err := gtk.FileChooserNativeDialogNew(
						"Select file to export",
						win,
						gtk.FILE_CHOOSER_ACTION_SAVE,
						"export",
						"cancel",
					)
					if err != nil {
						log.Fatal(err)
					}

					fc.SetCurrentFolder(CurrentFile.PathDir)
					fc.SetFilename(CurrentFile.Filename)

					switch fc.Run() {
					case int(gtk.RESPONSE_ACCEPT):
						name := fc.GetFilename()
						log.Println(name)
						returns <- name
					default:
						returns <- ""
					}
					fc.Destroy()

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

					fc.SetCurrentFolder(CurrentFile.PathDir)
					fc.SetFilename(CurrentFile.Filename)

					switch fc.Run() {
					case int(gtk.RESPONSE_ACCEPT):
						name := fc.GetFilename()
						log.Println(name)
						returns <- name
					default:
						returns <- ""
					}
					fc.Destroy()
				case "quit":
					running = false
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

func (s *UIControlSystem) process(component interface{}, isProcessingChildren bool) *Entity {
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
		return nil
	}

	if moveable.Bounds.Contains(rl.GetMousePosition().Subtract(moveable.Offset)) {
		hoverable.Hovered = true
		switch t := drawable.DrawableType.(type) {
		case *DrawableParent:
			for _, child := range t.Children {
				if r := s.process(child, true); r != nil {
					return r
				}
			}
		}

		// Scroll logic
		if scrollable != nil {
			scrollAmount := rl.GetMouseWheelMove()
			if scrollAmount != 0 {
				UIHasControl = true
				scrollable.ScrollOffset += scrollAmount
				return entity
			}
		}

		// Mouse button logic
		button := s.getButtonDown()

		if interactable != nil && drawable.Hidden == false {
			if button != MouseButtonNone {
				// Mouse events are handled by the caller function (Update)
				if interactable.OnMouseDown != nil || interactable.OnMouseUp != nil {
					interactable.ButtonDown = button
					// SetCapturedInput(entity, interactable)
					return entity
				}
			}
		}
	}

	return nil
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
				// log.Println(file)
				Files = append(Files, file)
				CurrentFile = file
				EditorsUIAddButton(file)
			}
		}
	}
}

func UIExport() {
	UIControlSystemCmds <- "export"
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

func (s *UIControlSystem) HandleKeyboardEvents() {
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

			// Prevent tool switching or anything that could alter the state of the tool being used
			// Moving the cursor with the keyboard is still allowed
			if rl.IsMouseButtonDown(rl.MouseLeftButton) || rl.IsMouseButtonDown(rl.MouseRightButton) || rl.IsMouseButtonDown(rl.MouseMiddleButton) {
				break
			}

			shouldReturn := true

			// Can work with entities which are capturing the input
			switch key {
			case "cancel":
				if UIInteractableCapturedInput != nil {
					// Escape from text entry
					// TODO
				} else {
					if CurrentFile.DoingSelection {
						CurrentFile.CancelSelection()
					}
				}
			case "delete":
				CurrentFile.DeleteSelection()
			case "copy":
				CurrentFile.Copy()
			case "paste":
				// Pixel paste
				CurrentFile.Paste()
				// Input paste
				if UIInteractableCapturedInput != nil && UIInteractableCapturedInput.OnKeyPress != nil {
					for _, char := range rl.GetClipboardText() {
						UIInteractableCapturedInput.OnKeyPress(UIEntityCapturedInput, rl.Key(char))
					}
				}
			default:
				shouldReturn = false
			}

			if shouldReturn {
				return
			}

			// Don't allow events to happen if a component is inputting text
			if UIEntityCapturedInput != nil {
				break
			}

			shouldReturn = true

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

			case "flipHorizontal":
				CurrentFile.FlipHorizontal()
			case "flipVertical":
				CurrentFile.FlipVertical()

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
			default:
				shouldReturn = false
			}

			if shouldReturn {
				return
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
}

func (s *UIControlSystem) Update(dt float32) {
	s.HandleKeyboardEvents()

	// Don't bother with UI controls, something is being drawn
	if FileHasControl {
		return
	}

	res := s.Scene.QueryTag(s.Scene.Tags["basic"], s.Scene.Tags["scrollable"], s.Scene.Tags["interactable"])
	// Reverse order so that entities that are on top can get input and return
	// TODO check if this actually matters or does anything
	for i := len(res)/2 - 1; i >= 0; i-- {
		opp := len(res) - 1 - i
		res[i], res[opp] = res[opp], res[i]
	}

	var entity *Entity
	UIHasControl = false

	// entity = UIEntityCapturedInput
	// the entity which would be returned from process()
	var newEntity *Entity
	for _, result := range res {
		newEntity = s.process(result, false)
		if newEntity != nil {
			break
		}
	}

	if UIEntityCapturedInput != nil {
		entity = UIEntityCapturedInput
	} else {
		entity = newEntity
	}

	button := s.getButtonDown()

	// Check if the UIEntityCapturedInput is the same as the newly clicked element
	if s.mouseButtonDown == false && button != MouseButtonNone {
		s.mouseButtonDown = true
		if UIEntityCapturedInput != newEntity {
			entity = newEntity
		}
	} else if s.mouseButtonDown == true && button == MouseButtonNone {
		s.mouseButtonDown = false
	}

	if entity != nil {
		UIHasControl = true
		if interactable, ok := entity.GetInteractable(); ok {
			// Continuously sends mouse down event
			lastButton := interactable.ButtonDown
			if lastButton != MouseButtonNone && button == lastButton {
				if interactable.ButtonReleased == true {
					interactable.ButtonDownAt = time.Now()
					interactable.ButtonReleased = false
				}

				isHeld := false
				if time.Now().Sub(interactable.ButtonDownAt) > time.Second/2 {
					isHeld = true
					if moveable, ok := entity.GetMoveable(); ok {
						if moveable.Draggable {
							UIIsDraggingEntity = true
						}
					}
				}

				if interactable.OnMouseDown != nil {
					interactable.OnMouseDown(UIEntityCapturedInput, lastButton, isHeld)
				}
			}

			// Only allow input capture to happen once per new entity
			if interactable.ButtonDown != MouseButtonNone && UIEntityCapturedInput != entity {
				SetCapturedInput(entity, interactable)
			}
		}
	}

	// Handle keyboard events
	if UIInteractableCapturedInput != nil && UIInteractableCapturedInput.OnKeyPress != nil {
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

		if entity != UIEntityCapturedInput && button != MouseButtonNone {
			RemoveCapturedInput()
		}
	} else if (UIEntityCapturedInput != nil || UIIsDraggingEntity) && button == MouseButtonNone {
		// Handle mouse up event
		if UIInteractableCapturedInput.ButtonReleased == false {
			UIInteractableCapturedInput.ButtonReleased = true
			if UIInteractableCapturedInput.OnMouseUp != nil {
				UIInteractableCapturedInput.OnMouseUp(UIEntityCapturedInput, UIInteractableCapturedInput.ButtonDown)
			}
			UIIsDraggingEntity = false
		}

		// Remove entity
		if UIInteractableCapturedInput.OnKeyPress == nil {
			RemoveCapturedInput()
			UIHasControl = false
		}
	}
}
