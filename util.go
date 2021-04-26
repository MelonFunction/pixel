package main

import (
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

func DiffUint8(a, b uint8, ratio float32) uint8 {
	if ratio == 0 {
		return a
	}
	return uint8(float32(a)*(1-ratio) + float32(b)*ratio)
}

// BlendWithOpacity assumes that b is the color being drawn on top
func BlendWithOpacity(a, b rl.Color) rl.Color {
	if b.A == 0 {
		return a
	}
	blendRatio := float32(a.A) / float32(b.A)
	// log.Println(float32(a.R), float32(b.R), blendRatio, DiffUint8(a.R, b.R, blendRatio))

	c := a
	// c.A = AddAndClampUint8(a.A, b.A)
	c.R = DiffUint8(a.R, b.R, blendRatio)
	c.G = DiffUint8(a.G, b.G, blendRatio)
	c.B = DiffUint8(a.B, b.B, blendRatio)
	return c
}
