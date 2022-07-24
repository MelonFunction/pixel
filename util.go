package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// IntVec2 is used mostly as a composite key for pixel data maps
type IntVec2 struct {
	X, Y int32
}

// MouseButton type
type MouseButton int32

// Key type
type Key int32

// Line draws pixels across a line (rl.DrawLine doesn't draw properly)
func Line(x0, y0, x1, y1 int32, drawFunc func(x, y int32)) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	var sx, sy int32
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	for {
		drawFunc(x0, y0)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// Rotate rotates v by phi
func (v IntVec2) Rotate(phi float64) IntVec2 {
	c, s := math.Cos(phi), math.Sin(phi)
	return IntVec2{int32(c*float64(v.X) - s*float64(v.Y)), int32(s*float64(v.X) + c*float64(v.Y))}
}

// AddAndClampUint8 adds two ints and caps them at uint8 max
func AddAndClampUint8(a, b uint8) uint8 {
	if int32(a)+int32(b) > 255 {
		return 255
	}
	return a + b
}

// BlendWithOpacity blends two colors together
// It assumes that b is the color being drawn on top
func BlendWithOpacity(a, b rl.Color) rl.Color {
	if b.A == 0 {
		return a
	}
	if a.A == 0 { // TODO check if a.A != 255 is better
		// return b
		a.A = 255
	}

	a.A = AddAndClampUint8(a.A, b.A/2)
	blendRatio := (float32(a.A) - float32(b.A)) / float32(a.A)

	c := rl.Color{
		A: a.A,
		R: uint8(float32(a.R)*blendRatio + float32(b.R)*(1-blendRatio)),
		G: uint8(float32(a.G)*blendRatio + float32(b.G)*(1-blendRatio)),
		B: uint8(float32(a.B)*blendRatio + float32(b.B)*(1-blendRatio)),
	}

	return c
}

// ColorToHex converts an rl.Color into a hex string
func ColorToHex(color rl.Color) string {
	return fmt.Sprintf("%02x%02x%02x%02x", color.R, color.G, color.B, color.A)
}

// HexToColor converts a hex string into a rl.Color
func HexToColor(color string) (rl.Color, error) {
	if len(color) > 0 {
		if color[0] == '#' {
			color = color[1:]
		}

		var r, g, b, a int64 = 0, 0, 0, 255
		var err error
		switch len(color) {
		case 8:
			if a, err = strconv.ParseInt(color[6:8], 16, 32); err != nil {
				log.Println(err)
			}
			fallthrough
		case 6:
			if r, err = strconv.ParseInt(color[0:2], 16, 32); err != nil {
				log.Println(err)
			}
			if g, err = strconv.ParseInt(color[2:4], 16, 32); err != nil {
				log.Println(err)
			}
			if b, err = strconv.ParseInt(color[4:6], 16, 32); err != nil {
				log.Println(err)
			}

			return rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a)), nil
		}
	}
	return rl.Color{}, fmt.Errorf("color couldn't be created from hex")
}

//go:embed res
var f embed.FS

// SetupFiles creates temp files from the embedded files in ./res
// It overwrites existing files
func SetupFiles() error {
	ex, err := os.UserCacheDir()
	if err != nil {
		log.Println(err)
		return err
	}
	savePath := path.Join(ex, "pixel")

	os.Mkdir(savePath, 0700)

	fs.WalkDir(f, "res", func(p string, d fs.DirEntry, err error) error {
		if data, err := os.ReadFile(p); err == nil {
			savePath := path.Join(savePath, p)
			nestedPath := filepath.Dir(savePath)

			os.MkdirAll(nestedPath, 0700)
			os.WriteFile(savePath, data, 0666)
		}

		return err
	})

	return nil
}

// GetFile will create a temp file for everything that was embedded in ./res,
// resPath is the relative path to the file
func GetFile(resPath string) string {
	ex, err := os.UserCacheDir()
	if err != nil {
		log.Println(err)
		return ""
	}

	cachePath := path.Join(ex, "pixel", resPath)
	_, err = os.Stat(cachePath)
	if err != nil {
		log.Println(err)
		return ""
	}

	return cachePath
}

// GetClampedCoordinates limits the x and y to the width/height of the canvas
func GetClampedCoordinates(x, y int32) IntVec2 {
	if x < 0 {
		x = 0
	} else if x >= CurrentFile.CanvasWidth-1 {
		x = CurrentFile.CanvasWidth - 1
	}
	if y < 0 {
		y = 0
	} else if y >= CurrentFile.CanvasHeight-1 {
		y = CurrentFile.CanvasHeight - 1
	}

	v := IntVec2{x, y}
	return v
}

// GetTilePosition returns the top left x and y coordinates
func GetTilePosition(x, y int32) IntVec2 {
	return IntVec2{
		X: x / CurrentFile.TileWidth * CurrentFile.TileWidth,
		Y: y / CurrentFile.TileHeight * CurrentFile.TileHeight,
	}
}

// MaxInt32 returns the bigger int32 of the two args
func MaxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// MinInt32 returns the smaller int32 of the two args
func MinInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
