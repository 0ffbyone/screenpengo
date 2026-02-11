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
	"screenpengo/internal/tool"
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

	// Draw all completed shapes.
	for i := range c.Shapes {
		r.renderShape(gtx.Ops, &c.Shapes[i])
	}

	// Draw current shape being drawn.
	if c.CurrentShape != nil {
		r.renderShape(gtx.Ops, c.CurrentShape)
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

// renderShape renders a geometric shape
func (r *GioRenderer) renderShape(ops *op.Ops, s *canvas.Shape) {
	strokeWidth := int(math.Max(2, float64(s.WidthPx)))

	switch s.Type {
	case tool.Circle:
		r.renderCircleShape(ops, s, strokeWidth)
	case tool.Rectangle:
		r.renderRectangleShape(ops, s, strokeWidth)
	case tool.Line:
		r.renderLineShape(ops, s, strokeWidth)
	case tool.Arrow:
		r.renderArrowShape(ops, s, strokeWidth)
	}
}

// renderCircleShape draws a circle outline (circumference only, not filled)
func (r *GioRenderer) renderCircleShape(ops *op.Ops, s *canvas.Shape, strokeWidth int) {
	// Calculate radius from start and end points
	dx := s.EndPos.X - s.StartPos.X
	dy := s.EndPos.Y - s.StartPos.Y
	radius := float64(math.Sqrt(float64(dx*dx + dy*dy)))

	if radius < 1 {
		return
	}

	centerX := s.StartPos.X
	centerY := s.StartPos.Y

	// Draw circle as a series of points around the circumference
	numPoints := int(math.Max(32, radius*2)) // More points for smoother circles
	circleRadius := int(math.Max(1, float64(strokeWidth/2)))

	for i := 0; i < numPoints; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(numPoints)
		x := int(centerX + float32(radius*math.Cos(angle)))
		y := int(centerY + float32(radius*math.Sin(angle)))

		// Draw a small circle at this point
		rect := image.Rect(x-circleRadius, y-circleRadius, x+circleRadius, y+circleRadius)
		paint.FillShape(ops, s.Color, clip.Ellipse(rect).Op(ops))
	}
}

// renderRectangleShape draws a rectangle outline
func (r *GioRenderer) renderRectangleShape(ops *op.Ops, s *canvas.Shape, strokeWidth int) {
	x1, y1 := int(s.StartPos.X), int(s.StartPos.Y)
	x2, y2 := int(s.EndPos.X), int(s.EndPos.Y)

	// Normalize coordinates
	if x2 < x1 {
		x1, x2 = x2, x1
	}
	if y2 < y1 {
		y1, y2 = y2, y1
	}

	// Draw four lines for the rectangle outline
	// Top
	r.renderThickLine(ops, image.Pt(x1, y1), image.Pt(x2, y1), strokeWidth, s.Color)
	// Right
	r.renderThickLine(ops, image.Pt(x2, y1), image.Pt(x2, y2), strokeWidth, s.Color)
	// Bottom
	r.renderThickLine(ops, image.Pt(x2, y2), image.Pt(x1, y2), strokeWidth, s.Color)
	// Left
	r.renderThickLine(ops, image.Pt(x1, y2), image.Pt(x1, y1), strokeWidth, s.Color)
}

// renderLineShape draws a straight line
func (r *GioRenderer) renderLineShape(ops *op.Ops, s *canvas.Shape, strokeWidth int) {
	r.renderThickLine(ops, image.Pt(int(s.StartPos.X), int(s.StartPos.Y)),
		image.Pt(int(s.EndPos.X), int(s.EndPos.Y)), strokeWidth, s.Color)
}

// renderArrowShape draws an arrow
func (r *GioRenderer) renderArrowShape(ops *op.Ops, s *canvas.Shape, strokeWidth int) {
	// Draw the main line
	r.renderThickLine(ops, image.Pt(int(s.StartPos.X), int(s.StartPos.Y)),
		image.Pt(int(s.EndPos.X), int(s.EndPos.Y)), strokeWidth, s.Color)

	// Calculate arrowhead
	dx := s.EndPos.X - s.StartPos.X
	dy := s.EndPos.Y - s.StartPos.Y
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if length < 1 {
		return
	}

	// Normalized direction
	dirX := dx / length
	dirY := dy / length

	// Arrowhead size (proportional to stroke width)
	arrowSize := float32(strokeWidth * 4)

	// Arrowhead points (60 degree angle)
	angle := float32(math.Pi / 6) // 30 degrees

	// Left wing
	cos1 := float32(math.Cos(float64(angle)))
	sin1 := float32(math.Sin(float64(angle)))
	leftX := s.EndPos.X - arrowSize*(dirX*cos1+dirY*sin1)
	leftY := s.EndPos.Y - arrowSize*(dirY*cos1-dirX*sin1)

	// Right wing
	rightX := s.EndPos.X - arrowSize*(dirX*cos1-dirY*sin1)
	rightY := s.EndPos.Y - arrowSize*(dirY*cos1+dirX*sin1)

	// Draw arrowhead wings
	r.renderThickLine(ops, image.Pt(int(s.EndPos.X), int(s.EndPos.Y)),
		image.Pt(int(leftX), int(leftY)), strokeWidth, s.Color)
	r.renderThickLine(ops, image.Pt(int(s.EndPos.X), int(s.EndPos.Y)),
		image.Pt(int(rightX), int(rightY)), strokeWidth, s.Color)
}

// renderThickLine draws a line with thickness by drawing circles along the path
func (r *GioRenderer) renderThickLine(ops *op.Ops, start, end image.Point, thickness int, col color.NRGBA) {
	radius := int(math.Max(1, float64(thickness/2)))

	// Calculate number of points based on line length
	dx := float64(end.X - start.X)
	dy := float64(end.Y - start.Y)
	length := math.Sqrt(dx*dx + dy*dy)
	steps := int(math.Max(2, length/float64(radius)))

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(float64(start.X)*(1-t) + float64(end.X)*t)
		y := int(float64(start.Y)*(1-t) + float64(end.Y)*t)

		rect := image.Rect(x-radius, y-radius, x+radius, y+radius)
		paint.FillShape(ops, col, clip.Ellipse(rect).Op(ops))
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
