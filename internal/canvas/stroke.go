package canvas

import (
	"image/color"
	"math"

	"gioui.org/f32"
)

type Stroke struct {
	Points []f32.Point
	Color  color.NRGBA
	Width  float32
}

func appendInterpolated(dst *[]f32.Point, a, b f32.Point, spacing float32) {
	if spacing <= 1 {
		*dst = append(*dst, b)
		return
	}
	dx := float64(b.X - a.X)
	dy := float64(b.Y - a.Y)
	d := math.Hypot(dx, dy)
	if d == 0 {
		return
	}
	steps := int(d / float64(spacing))
	if steps < 1 {
		*dst = append(*dst, b)
		return
	}
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		p := f32.Point{
			X: float32(float64(a.X) + dx*t),
			Y: float32(float64(a.Y) + dy*t),
		}
		*dst = append(*dst, p)
	}
}
