package main

import (
	"log"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	// the buttons themselves
	menuButtons *Entity
)

// NewMenuUI returns a new entity
func NewMenuUI(bounds rl.Rectangle) *Entity {
	// Top level dropdown buttons
	var fileButton, editButton, paletteButton *Entity
	// submenus
	var fileSubMenu, editSubMenu, paletteSubMenu *Entity

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
	measured = rl.MeasureTextEx(Font, " file ", UIFontSize, 1)
	fileButton = NewButtonText(
		rl.NewRectangle(100, 100, measured.X+10, UIFontSize*2),
		" file ", TextAlignCenter, false, func(entity *Entity, button MouseButton) {
			showDropdown(entity, fileSubMenu)
		}, nil)

	measured = rl.MeasureTextEx(Font, " edit ", UIFontSize, 1)
	editButton = NewButtonText(
		rl.NewRectangle(100, 100, measured.X+10, UIFontSize*2),
		" edit ", TextAlignCenter, false, func(entity *Entity, button MouseButton) {
			showDropdown(entity, editSubMenu)
		}, nil)

	measured = rl.MeasureTextEx(Font, " palette ", UIFontSize, 1)
	paletteButton = NewButtonText(
		rl.NewRectangle(100, 100, measured.X+10, UIFontSize*2),
		" palette ", TextAlignCenter, false, func(entity *Entity, button MouseButton) {
			showDropdown(entity, paletteSubMenu)
		}, nil)

	// Add to the bar
	menuButtons = NewBox(bounds, []*Entity{
		fileButton,
		editButton,
		paletteButton,
	}, FlowDirectionHorizontal)
	menuButtons.FlowChildren()

	// File menu
	measured = rl.MeasureTextEx(Font, "close file    ", UIFontSize, 1)
	bounds.Y += UIFontSize * 2
	bounds.Height = float32(rl.GetScreenHeight())
	bounds.Width = measured.X + 10
	fileSubMenu = NewBox(bounds, []*Entity{
		NewButtonText( // New
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"new", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				UINew()
			}, nil),
		NewButtonText( // Save
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"save", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				if len(CurrentFile.FileDir) > 0 {
					CurrentFile.SaveAs(CurrentFile.FileDir)
				} else {
					UISaveAs()
				}
			}, nil),
		NewButtonText( // Save As
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"save as", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				UISaveAs()
			}, nil),
		NewButtonText( // Open
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"open", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				UIOpen()
			}, nil),
		NewButtonText( // Close
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"close file", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				UIClose()
			}, nil),
		NewButtonText( // Resize
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"resize", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				ResizeUIShowDialog()
			}, nil),
	}, FlowDirectionVertical)
	fileSubMenu.FlowChildren()
	fileSubMenu.Hide()

	// Edit menu
	measured = rl.MeasureTextEx(Font, "flip (horizontal) ", UIFontSize, 1)
	fileButtonMoveable, ok := editButton.GetMoveable()
	if !ok {
		log.Panic("fileButton error")
	}
	bounds.X += fileButtonMoveable.Bounds.Width
	bounds.Width = measured.X + 10
	editSubMenu = NewBox(bounds, []*Entity{
		NewButtonText( // Flip (horizontal)
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"flip (horizontal)", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				CurrentFile.FlipHorizontal()
			}, nil),
		NewButtonText( // Flip (vertical)
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"flip (vertical)", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				CurrentFile.FlipVertical()
			}, nil),
		NewButtonText( // Outline
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"outline", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				CurrentFile.Outline()
			}, nil),
	}, FlowDirectionVertical)
	editSubMenu.FlowChildren()
	editSubMenu.Hide()

	// Palette menu
	measured = rl.MeasureTextEx(Font, "delete (hold shift) ", UIFontSize, 1)
	editButtonMoveable, ok := editButton.GetMoveable()
	if !ok {
		log.Panic("editButton error")
	}
	bounds.X += editButtonMoveable.Bounds.Width
	bounds.Width = measured.X + 10
	paletteSubMenu = NewScrollableList(bounds, []*Entity{
		NewButtonText( // New
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"new", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				Settings.PaletteData = append(Settings.PaletteData, Palette{
					Name: "new",
				})
				currentPalette := len(Settings.PaletteData) - 1
				CurrentFile.CurrentPalette = int32(currentPalette)
				SaveSettings()

				PaletteUIRebuildPalette()
				paletteSubMenu.Hide()
			}, nil),
		NewButtonText( // Delete
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"delete (hold shift)", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				if (rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)) && CurrentFile.CurrentPalette != 0 {
					Settings.PaletteData = append(Settings.PaletteData[:CurrentFile.CurrentPalette], Settings.PaletteData[CurrentFile.CurrentPalette+1:]...)
					CurrentFile.CurrentPalette = 0
					SaveSettings()

					PaletteUIRebuildPalette()
					paletteSubMenu.Hide()
				}
			}, nil),
		NewButtonText( // Duplicate
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"duplicate", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				Settings.PaletteData = append(Settings.PaletteData, Settings.PaletteData[CurrentFile.CurrentPalette])
				currentPalette := len(Settings.PaletteData) - 1
				CurrentFile.CurrentPalette = int32(currentPalette)
				Settings.PaletteData[currentPalette].Name += "(1)"
				SaveSettings()

				PaletteUIRebuildPalette()
				paletteSubMenu.Hide()
			}, nil),
		NewButtonText( // Create From Image
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"create from image", TextAlignLeft, false, func(entity *Entity, button MouseButton) {
				colors := make(map[rl.Color]struct{})
				colorsSlice := make([]rl.Color, 0)
				cl := CurrentFile.GetCurrentLayer().PixelData
				for x := int32(0); x < CurrentFile.CanvasWidth; x++ {
					for y := int32(0); y < CurrentFile.CanvasHeight; y++ {
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
				CurrentFile.CurrentPalette = int32(currentPalette)
				SaveSettings()

				PaletteUIRebuildPalette()
				paletteSubMenu.Hide()
			}, nil),
		NewButtonText( // Load Items Spacer
			rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
			"---- Load ----", TextAlignCenter, false, func(entity *Entity, button MouseButton) {
			}, nil),
	}, FlowDirectionVertical)
	paletteSubMenu.FlowChildren()
	paletteSubMenu.Hide()

	if drawable, ok := paletteSubMenu.GetDrawable(); ok {
		var originalChildrenLen int32
		if children, err := paletteSubMenu.GetChildren(); err == nil {
			originalChildrenLen = int32(len(children))
		} else {
			log.Panic(err)
		}

		var alreadyShowing bool

		drawable.OnShow = func(entity *Entity) {
			// add an entry for every palette available to be loaded
			if alreadyShowing {
				return
			}
			alreadyShowing = true
			for i, palette := range Settings.PaletteData {
				p := i
				paletteSubMenu.PushChild(
					NewButtonText(
						rl.NewRectangle(0, 0, measured.X+10, UIFontSize*2),
						palette.Name, TextAlignLeft, false, func(entity *Entity, button MouseButton) {
							CurrentFile.CurrentPalette = int32(p)
							PaletteUIRebuildPalette()
						}, nil))
			}
			paletteSubMenu.FlowChildren()
		}
		drawable.OnHide = func(entity *Entity) {
			alreadyShowing = false
			// clear away the available palettes
			if children, err := paletteSubMenu.GetChildren(); err == nil {
				for i := int32(len(children) - 1); i >= originalChildrenLen; i-- {
					paletteSubMenu.RemoveChild(children[i])
				}
			} else {
				log.Panic(err)
			}
		}
	}

	return menuButtons
}
