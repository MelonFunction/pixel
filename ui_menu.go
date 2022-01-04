package main

import (
	"log"
	"sync"
	"time"

	rl "github.com/lachee/raylib-goplus/raylib"
)

var (
	// the buttons themselves
	menuButtons *Entity
)

// NewMenuUI returns a new entity
func NewMenuUI(bounds rl.Rectangle) *Entity {
	// Top level dropdown buttons
	var fileButton, paletteButton *Entity
	// fileButton buttons
	var newButton, saveButton, saveAsButton, openButton, resizeButton *Entity
	// paletteButton buttons
	var newPaletteButton, deletePaletteButton, duplicatePaletteButton, canvasToPaletteButton, spacerPaletteButton *Entity
	// submenus
	var fileSubMenu, paletteSubMenu *Entity

	// button is top level menu button, dropdown is the child elements,
	showDropdown := func(button *Entity, dropdown *Entity) {
		dropdown.Show()
		mutex := &sync.RWMutex{}

		// Clicked on button should simulate being hovered
		hovered := map[*Entity]struct{}{
			button: {},
		}

		handleHovered := func(entity *Entity) {

			if hoverable, ok := entity.GetHoverable(); ok {
				hoverable.OnMouseEnter = func(entity *Entity) {
					mutex.Lock()
					hovered[entity] = struct{}{}
					mutex.Unlock()
				}
				hoverable.OnMouseLeave = func(entity *Entity) {
					go func() {
						mutex.Lock()
						delete(hovered, entity)
						mutex.Unlock()

						time.Sleep(500 * time.Millisecond)
						mutex.Lock()
						if len(hovered) == 0 {
							if scrollable, ok := dropdown.GetScrollable(); ok {
								scrollable.ScrollOffset = 0
							}
							dropdown.Hide()
						}
						mutex.Unlock()
					}()
				}
			}
		}

		handleHovered(button)
		if children, err := dropdown.GetChildren(); err == nil {
			for _, child := range children {
				handleHovered(child)
			}
		}
	}

	// Parent buttons
	var measured rl.Vector2
	measured = rl.MeasureTextEx(*Font, " file ", UIFontSize, 1)
	fileButton = NewButtonText(
		rl.NewRectangle(100, 100, measured.X+10, UIFontSize*2),
		" file ", TextAlignCenter, false, func(entity *Entity, button rl.MouseButton) {
			showDropdown(entity, fileSubMenu)
		}, nil)

	measured = rl.MeasureTextEx(*Font, " palette ", UIFontSize, 1)
	paletteButton = NewButtonText(
		rl.NewRectangle(100, 100, measured.X+10, UIFontSize*2),
		" palette ", TextAlignCenter, false, func(entity *Entity, button rl.MouseButton) {
			showDropdown(entity, paletteSubMenu)
		}, nil)

	// Add to the bar
	menuButtons = NewBox(bounds, []*Entity{
		fileButton,
		paletteButton,
	}, FlowDirectionHorizontal)

	menuButtons.FlowChildren()

	//
	// fileButton contents
	//

	measured = rl.MeasureTextEx(*Font, "save as ", UIFontSize, 1)

	newButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"new", TextAlignLeft, false, func(entity *Entity, button rl.MouseButton) {
			UINew()
		}, nil)

	saveButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"save", TextAlignLeft, false, func(entity *Entity, button rl.MouseButton) {
			if len(CurrentFile.FileDir) > 0 {
				CurrentFile.SaveAs(CurrentFile.FileDir)
			} else {
				UISaveAs()
			}
		}, nil)

	saveAsButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"save as", TextAlignLeft, false, func(entity *Entity, button rl.MouseButton) {
			UISaveAs()
		}, nil)

	openButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"open", TextAlignLeft, false, func(entity *Entity, button rl.MouseButton) {
			UIOpen()
		}, nil)

	resizeButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"resize", TextAlignLeft, false, func(entity *Entity, button rl.MouseButton) {
			ResizeUIShowDialog()
		}, nil)

	// File menu
	bounds.Y += UIFontSize * 2
	bounds.Height = float32(rl.GetScreenHeight())
	bounds.Width = measured.X + 10
	fileSubMenu = NewBox(bounds, []*Entity{
		newButton,
		saveButton,
		saveAsButton,
		openButton,
		resizeButton,
	}, FlowDirectionVertical)
	fileSubMenu.FlowChildren()
	fileSubMenu.Hide()

	//
	// paletteButton contents
	//
	measured = rl.MeasureTextEx(*Font, "delete (hold shift) ", UIFontSize, 1)

	newPaletteButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"new", TextAlignLeft, false, func(entity *Entity, button rl.MouseButton) {
			Settings.PaletteData = append(Settings.PaletteData, Palette{
				Name: "new",
			})
			currentPalette := len(Settings.PaletteData) - 1
			CurrentFile.CurrentPalette = currentPalette
			SaveSettings()

			PaletteUIRebuildPalette()
			paletteSubMenu.Hide()
		}, nil)

	deletePaletteButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"delete (hold shift)", TextAlignLeft, false, func(entity *Entity, button rl.MouseButton) {
			if (rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)) && CurrentFile.CurrentPalette != 0 {
				Settings.PaletteData = append(Settings.PaletteData[:CurrentFile.CurrentPalette], Settings.PaletteData[CurrentFile.CurrentPalette+1:]...)
				CurrentFile.CurrentPalette = 0
				SaveSettings()

				PaletteUIRebuildPalette()
				paletteSubMenu.Hide()
			}
		}, nil)

	duplicatePaletteButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"duplicate", TextAlignLeft, false, func(entity *Entity, button rl.MouseButton) {
			Settings.PaletteData = append(Settings.PaletteData, Settings.PaletteData[CurrentFile.CurrentPalette])
			currentPalette := len(Settings.PaletteData) - 1
			CurrentFile.CurrentPalette = currentPalette
			Settings.PaletteData[currentPalette].Name += "(1)"
			SaveSettings()

			PaletteUIRebuildPalette()
			paletteSubMenu.Hide()
		}, nil)

	canvasToPaletteButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"create from image", TextAlignLeft, false, func(entity *Entity, button rl.MouseButton) {
			colors := make(map[rl.Color]struct{})
			colorsSlice := make([]rl.Color, 0)
			cl := CurrentFile.GetCurrentLayer().PixelData
			for x := 0; x < CurrentFile.CanvasWidth; x++ {
				for y := 0; y < CurrentFile.CanvasHeight; y++ {
					color := cl[IntVec2{x, y}]
					if _, ok := colors[color]; !ok {
						colorsSlice = append(colorsSlice, color)
						colors[color] = struct{}{}
					}
				}
			}

			Settings.PaletteData = append(Settings.PaletteData, Palette{
				Name: "new",
				data: colorsSlice,
			})
			currentPalette := len(Settings.PaletteData) - 1
			CurrentFile.CurrentPalette = currentPalette
			SaveSettings()

			PaletteUIRebuildPalette()
			paletteSubMenu.Hide()
		}, nil)

	spacerPaletteButton = NewButtonText(
		rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
		"---- Load ----", TextAlignCenter, false, func(entity *Entity, button rl.MouseButton) {
		}, nil)

	// Palette menu
	fileButtonMoveable, ok := fileButton.GetMoveable()
	if !ok {
		log.Panic("fileButton error")
	}
	bounds.X += fileButtonMoveable.Bounds.Width
	bounds.Width = measured.X + 10
	paletteSubMenu = NewScrollableList(bounds, []*Entity{
		newPaletteButton,
		deletePaletteButton,
		duplicatePaletteButton,
		canvasToPaletteButton,
		spacerPaletteButton,
	}, FlowDirectionVertical)
	paletteSubMenu.FlowChildren()
	paletteSubMenu.Hide()
	if drawable, ok := paletteSubMenu.GetDrawable(); ok {
		var originalChildrenLen int
		if children, err := paletteSubMenu.GetChildren(); err == nil {
			originalChildrenLen = len(children)
		} else {
			log.Panic(err)
		}

		drawable.OnShow = func(entity *Entity) {
			// add an entry for every palette available to be loaded
			for i, palette := range Settings.PaletteData {
				p := i
				paletteSubMenu.PushChild(
					NewButtonText(
						rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
						palette.Name, TextAlignLeft, false, func(entity *Entity, button rl.MouseButton) {
							CurrentFile.CurrentPalette = p
							PaletteUIRebuildPalette()
						}, nil))
			}
			paletteSubMenu.FlowChildren()
		}
		drawable.OnHide = func(entity *Entity) {
			// clear away the available palettes
			if children, err := paletteSubMenu.GetChildren(); err == nil {
				for i := len(children) - 1; i >= originalChildrenLen; i-- {
					paletteSubMenu.RemoveChild(children[i])
				}
			} else {
				log.Panic(err)
			}
		}
	}

	return menuButtons
}
