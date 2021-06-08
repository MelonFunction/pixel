package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"

	rl "github.com/lachee/raylib-goplus/raylib"
)

// IntVec2 is used mostly as a composite key for pixel data maps
type IntVec2 struct {
	X, Y int
}

// Line draws pixels across a line (rl.DrawLine doesn't draw properly)
func Line(x0, y0, x1, y1 int, drawFunc func(x, y int)) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	var sx, sy int
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

func AddAndClampUint8(a, b uint8) uint8 {
	if int(a)+int(b) > 255 {
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
	if a.A == 0 {
		return b
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

		return rl.Color{uint8(r), uint8(g), uint8(b), uint8(a)}, nil
	}
	return rl.Color{}, fmt.Errorf("color couldn't be created from hex")
}

//go:embed res
var f embed.FS

// SetupFiles creates temp files from the embedded files in ./res
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

			if _, err := os.Stat(savePath); os.IsNotExist(err) {
				log.Println("Creating cache file: ", savePath)
				os.MkdirAll(nestedPath, 0700)
				os.WriteFile(savePath, data, 0666)
			}
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
