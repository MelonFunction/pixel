package main

import (
	"testing"

	rl "github.com/lachee/raylib-goplus/raylib"
)

func TestBlendWithOpacity(t *testing.T) {
	a := rl.NewColor(255, 0, 0, 128)
	b := rl.NewColor(255, 0, 0, 128)

	c := BlendWithOpacity(a, b)
	if c.A != 255 {
		t.Errorf("Alpha not 255")
	}
}
