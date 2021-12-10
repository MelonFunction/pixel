Refactor
  - system_file, system_render, system_controls don't have clear goals since system_file also renders.
    This should be moved to system_render. 
    
UI
  - Resize element
  - Double click to change text in label, single click to activate it

System
  - Dropping a file on the window should open it
  //- Serialize state data into file (like .ase etc)
    //- Animations
    //- Layers

Testing
  - lol

Previews (including layers)
  - non-square canvas shouldn't squish
  - allow movement/zoom
  
Layers
  Actions
    //- Merge down
    //- Delete
    //- Move up
    //- Move down
    //- Hide/show

Menubar
  Palettes
    //- Maybe show the palette name in the menu item?
    //- Save palette
    //- Load palette
    //- Rename palette
    //- Create palette from colors in image

Tools
  Shade Brush

  Pixel Brush
    //- Hold button to draw line
 
  Selector
    - CTRL+A should select everything (and switch tool to selector)
    - Draw in selection/mask
    - Rotate
    //- Copy
    //- Paste
    //- Copy/paste while moving
    //- Resize
    //- Flip
      //- rl.KeyH for Horizontal
      //- rl.KeyV for Vertical

Features
  Animation Tab
    //- Create animation button
    //- List (like layers)
      //- When clicked, allow tiles to be selected
    - Export
      - Name
      - Frames
      - Delays
  
  Tiles Tab
    actually, maybe the current spritesheet could just be extended and used like
    how the workflow would be in asesprite. maybe a tile/selection stamping tool
    could be made
    - Tiles from the tilemap can be placed in a grid
    - Each tile placed will map to a location on the spritesheet
    - Export
  
  Preview Panel
    //- Show the animation if the animation tab is selected
      //- Animation speed
      //- Pause/play
    //- Show the tile/map editor/placer if the tiles tab is selected
      //- Show the current tile being tiled in all directions
      - Allow editing in preview panel