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

type GioRenderer struct {
	Dim bool
}

func (r *GioRenderer) RenderFrame(gtx layout.Context, c *canvas.Canvas, cursorPos image.Point, cursorRadius int, showCursor bool) {
	paint.FillShape(gtx.Ops, color.NRGBA{A: 0}, clip.Rect{Max: gtx.Constraints.Max}.Op())

	if r.Dim {
		paint.FillShape(gtx.Ops, color.NRGBA{A: 120}, clip.Rect{Max: gtx.Constraints.Max}.Op())
	}

	for i := range c.Strokes {
		r.renderStroke(gtx.Ops, &c.Strokes[i])
	}

	if c.Current != nil {
		r.renderStroke(gtx.Ops, c.Current)
	}

	for i := range c.Shapes {
		r.renderShape(gtx.Ops, &c.Shapes[i])
	}

	if c.CurrentShape != nil {
		r.renderShape(gtx.Ops, c.CurrentShape)
	}

	if showCursor && cursorRadius > 0 {
		r.renderCursor(gtx.Ops, cursorPos, cursorRadius)
	}
}

func (r *GioRenderer) renderCursor(ops *op.Ops, pos image.Point, radius int) {
	rect := image.Rect(pos.X-radius, pos.Y-radius, pos.X+radius, pos.Y+radius)

	paint.FillShape(ops, color.NRGBA{R: 0, G: 0, B: 0, A: 150}, clip.Ellipse(rect).Op(ops))

	innerRadius := int(math.Max(1, float64(radius-2)))
	innerRect := image.Rect(pos.X-innerRadius, pos.Y-innerRadius, pos.X+innerRadius, pos.Y+innerRadius)
	paint.FillShape(ops, color.NRGBA{A: 0}, clip.Ellipse(innerRect).Op(ops))
}

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

func (r *GioRenderer) renderCircleShape(ops *op.Ops, s *canvas.Shape, strokeWidth int) {
	dx := s.EndPos.X - s.StartPos.X
	dy := s.EndPos.Y - s.StartPos.Y
	radius := float64(math.Sqrt(float64(dx*dx + dy*dy)))

	if radius < 1 {
		return
	}

	centerX := s.StartPos.X
	centerY := s.StartPos.Y

	numPoints := int(math.Max(32, radius*2))
	circleRadius := int(math.Max(1, float64(strokeWidth/2)))

	for i := 0; i < numPoints; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(numPoints)
		x := int(centerX + float32(radius*math.Cos(angle)))
		y := int(centerY + float32(radius*math.Sin(angle)))

		rect := image.Rect(x-circleRadius, y-circleRadius, x+circleRadius, y+circleRadius)
		paint.FillShape(ops, s.Color, clip.Ellipse(rect).Op(ops))
	}
}

func (r *GioRenderer) renderRectangleShape(ops *op.Ops, s *canvas.Shape, strokeWidth int) {
	x1, y1 := int(s.StartPos.X), int(s.StartPos.Y)
	x2, y2 := int(s.EndPos.X), int(s.EndPos.Y)

	if x2 < x1 {
		x1, x2 = x2, x1
	}
	if y2 < y1 {
		y1, y2 = y2, y1
	}

	r.renderThickLine(ops, image.Pt(x1, y1), image.Pt(x2, y1), strokeWidth, s.Color)
	r.renderThickLine(ops, image.Pt(x2, y1), image.Pt(x2, y2), strokeWidth, s.Color)
	r.renderThickLine(ops, image.Pt(x2, y2), image.Pt(x1, y2), strokeWidth, s.Color)
	r.renderThickLine(ops, image.Pt(x1, y2), image.Pt(x1, y1), strokeWidth, s.Color)
}

func (r *GioRenderer) renderLineShape(ops *op.Ops, s *canvas.Shape, strokeWidth int) {
	r.renderThickLine(ops, image.Pt(int(s.StartPos.X), int(s.StartPos.Y)),
		image.Pt(int(s.EndPos.X), int(s.EndPos.Y)), strokeWidth, s.Color)
}

func (r *GioRenderer) renderArrowShape(ops *op.Ops, s *canvas.Shape, strokeWidth int) {
	r.renderThickLine(ops, image.Pt(int(s.StartPos.X), int(s.StartPos.Y)),
		image.Pt(int(s.EndPos.X), int(s.EndPos.Y)), strokeWidth, s.Color)

	dx := s.EndPos.X - s.StartPos.X
	dy := s.EndPos.Y - s.StartPos.Y
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if length < 1 {
		return
	}

	dirX := dx / length
	dirY := dy / length

	arrowSize := float32(strokeWidth * 4)

	angle := float32(math.Pi / 6)

	cos1 := float32(math.Cos(float64(angle)))
	sin1 := float32(math.Sin(float64(angle)))
	leftX := s.EndPos.X - arrowSize*(dirX*cos1+dirY*sin1)
	leftY := s.EndPos.Y - arrowSize*(dirY*cos1-dirX*sin1)

	rightX := s.EndPos.X - arrowSize*(dirX*cos1-dirY*sin1)
	rightY := s.EndPos.Y - arrowSize*(dirY*cos1+dirX*sin1)

	r.renderThickLine(ops, image.Pt(int(s.EndPos.X), int(s.EndPos.Y)),
		image.Pt(int(leftX), int(leftY)), strokeWidth, s.Color)
	r.renderThickLine(ops, image.Pt(int(s.EndPos.X), int(s.EndPos.Y)),
		image.Pt(int(rightX), int(rightY)), strokeWidth, s.Color)
}

func (r *GioRenderer) renderThickLine(ops *op.Ops, start, end image.Point, thickness int, col color.NRGBA) {
	radius := int(math.Max(1, float64(thickness/2)))

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
