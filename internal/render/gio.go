package render

import (
	"image"
	"image/color"
	"math"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"

	"screenpengo/internal/canvas"
)

// GioRenderer handles rendering strokes using the Gio framework.
type GioRenderer struct {
	Dim bool
}

// RenderFrame renders the complete frame including background, dimming, all strokes, and cursor.
func (r *GioRenderer) RenderFrame(gtx layout.Context, c *canvas.Canvas, cursorPos image.Point, cursorRadius int, showCursor bool) {
	// Background (transparent).
	paint.FillShape(gtx.Ops, color.NRGBA{A: 0}, clip.Rect{Max: gtx.Constraints.Max}.Op())

	// Dimming overlay if enabled.
	if r.Dim {
		paint.FillShape(gtx.Ops, color.NRGBA{A: 120}, clip.Rect{Max: gtx.Constraints.Max}.Op())
	}

	// Draw all completed strokes.
	for i := range c.Strokes {
		r.renderStroke(gtx.Ops, &c.Strokes[i])
	}

	// Draw current stroke being drawn.
	if c.Current != nil {
		r.renderStroke(gtx.Ops, c.Current)
	}

	// Draw cursor circle to show brush size
	if showCursor && cursorRadius > 0 {
		r.renderCursor(gtx.Ops, cursorPos, cursorRadius)
	}
}

// renderCursor draws a circle outline at the cursor position to show brush size
func (r *GioRenderer) renderCursor(ops *op.Ops, pos image.Point, radius int) {
	// Draw a semi-transparent circle outline
	rect := image.Rect(pos.X-radius, pos.Y-radius, pos.X+radius, pos.Y+radius)

	// Outer circle (outline)
	paint.FillShape(ops, color.NRGBA{R: 0, G: 0, B: 0, A: 150}, clip.Ellipse(rect).Op(ops))

	// Inner circle (to create outline effect)
	innerRadius := int(math.Max(1, float64(radius-2)))
	innerRect := image.Rect(pos.X-innerRadius, pos.Y-innerRadius, pos.X+innerRadius, pos.Y+innerRadius)
	paint.FillShape(ops, color.NRGBA{A: 0}, clip.Ellipse(innerRect).Op(ops))
}

// renderStroke renders a single stroke as a series of filled circles.
func (r *GioRenderer) renderStroke(ops *op.Ops, s *canvas.Stroke) {
	if len(s.Points) == 0 {
		return
	}
	radius := int(math.Max(1, float64(s.Width/2)))
	for _, p := range s.Points {
		rect := image.Rect(int(p.X)-radius, int(p.Y)-radius, int(p.X)+radius, int(p.Y)+radius)
		paint.FillShape(ops, s.Color, clip.Ellipse(rect).Op(ops))
	}
}
