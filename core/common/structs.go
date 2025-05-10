package com

type Rgb struct {
	R int
	G int
	B int
}

func NewRgb(r, g, b int) Rgb {
	return Rgb{
		R: r,
		G: g,
		B: b,
	}
}

func NewRgb64(r, g, b int64) Rgb {
	return NewRgb(int(r), int(g), int(b))
}
