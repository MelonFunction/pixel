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
