package canvas

import (
	"image/color"

	"gioui.org/f32"
)

// Canvas manages all drawing strokes.
type Canvas struct {
	Strokes []Stroke
	Current *Stroke
}

// StartStroke begins a new stroke with the given color and width.
func (c *Canvas) StartStroke(color color.NRGBA, widthPx float32, startPoint f32.Point) {
	c.Current = &Stroke{
		Color:  color,
		Width:  widthPx,
		Points: []f32.Point{startPoint},
	}
}

// AddPoint adds a point to the current stroke with interpolation for smoothness.
func (c *Canvas) AddPoint(point f32.Point) {
	if c.Current == nil {
		return
	}
	last := c.Current.Points[len(c.Current.Points)-1]
	appendInterpolated(&c.Current.Points, last, point, c.Current.Width/2)
}

// FinishStroke commits the current stroke to the canvas.
func (c *Canvas) FinishStroke() {
	if c.Current != nil {
		c.Strokes = append(c.Strokes, *c.Current)
		c.Current = nil
	}
}

// Clear removes all strokes from the canvas.
func (c *Canvas) Clear() {
	c.Strokes = nil
	c.Current = nil
}
