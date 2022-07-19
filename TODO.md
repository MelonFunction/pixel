🔴 Not started
🟠 Started
🔵 Has issues or testing needed
🟢 Done

Refactor
  🔴 system_file, system_render, system_controls don't have clear goals since system_file also renders.
    This should be moved to system_render. 
    
UI
  🔴 Resize element
  🔴 Double click to change text in label, single click to activate it

System
  🔴 * on unsaved files
  🔴 Prevent quit when there is an unsaved file
  🔴 Opening image with transparency causes colored artifacts
    🔴 Also transperency is wrong. Probs blending function
  🔴 Dropping a file on the window should open it
  🟢 Serialize state data into file (like .ase etc)
    🟢 Animations
    🟢 Layers

Testing
  🔴 lol

Previews (including layers)
  🟢 maintain aspect ratios
  🔴 allow movement/zoom
  🔴 lock preview window position to cell (with a hotkey)
  
Layers
  Actions
    🟢 Merge down
    🟢 Delete
    🟢 Move up
    🟢 Move down
    🟢 Hide/show

Palettes
  🔴 Hold shift to change the "add color to palette (+) button" to "remove color from palette (-) button"
  🟢 Highlight left/right color after click (un-highlight if color adjusted with controls)

Menubar
  Palettes
    🟢 Maybe show the palette name in the menu item?
    🟢 Save palette
    🟢 Load palette
    🟢 Rename palette
    🟢 Create palette from colors in image

Tools
  🔴 Shade Brush

  Pixel Brush
    🟢 Hold button to draw line
 
  Selector
    🟢 Resize should flip selection
    🔴 CTRL+A should select everything (and switch tool to selector)
    🔴 Draw in selection/mask
    🔴 Rotate
    🔴 Resize UI controls should have handles larger than 1px
    🟢 Copy
    🟢 Paste
    🟢 Copy/paste while moving
    🟢 Resize
    🟢 Flip
      🟢 rl.KeyH for Horizontal
      🟢 rl.KeyV for Vertical

Features
  Animation Tab
    🟢 Create animation button
    🟢 List (like layers)
      🟢 When clicked, allow tiles to be selected
    🔴 Export
      🔴 Name
      🔴 Frames
      🔴 Delays
  
  Preview Panel
    🟢 Show the animation if the animation tab is selected
      🟢 Animation speed
      🟢 Pause/play
    🟢 Show the tile/map editor/placer if the tiles tab is selected
      🟢 Show the current tile being tiled in all directions
      🔴 Allow editing in preview panel