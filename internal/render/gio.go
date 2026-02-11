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

// RenderFrame renders the complete frame including background, dimming, and all strokes.
func (r *GioRenderer) RenderFrame(gtx layout.Context, c *canvas.Canvas) {
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
