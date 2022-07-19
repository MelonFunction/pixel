# MelonPixel

🚧 Under heavy development! 🚧

⚠️ There are some serious problems which I'm working on! Check the [todo list](TODO.txt) ⚠️

## Features
- Tabbed files
- Palettes
    - Multiple palettes supported
    - Change color with the keyboard
    - Add and remove colors easily
- History (undo/redo for every action)
- Tools/Operations:
    - Pencil/eraser/brush 
        - Changeable size
    - Fill
    - Color picker
    - Selection (rectangle selection only currently)
    - Flip selection (or the entire canvas if there isn't a selection)
    - Move and resize the selection
    - Outline the selection (or the entire canvas there isn't a selection)
- Color picker
    - Updates indicator position when a palette color is selected
    - Alpha slider
- Preview
    - Full canvas view
    - Repeating tile view
    - Zoomed view
    - Animation view
- Animation
    - Create basic animations
    - Select tiles to be in the animation
    - Fixed frame time (complex animations are beyond the scope of this program)
- Control the cursor with the keyboard
- Layers
    - Hide
    - Move up or down
    - Merge with the layer below
- Resize canvas and tile size easily

## Installation
```
git clone github.com/MelonFunction/pixel
cd pixel
go get -v ./...
go run .
```

## Dependencies:  
Install whatever these libraries say to install!
- https://github.com/gotk3/gotk3
- https://github.com/gen2brain/raylib-go