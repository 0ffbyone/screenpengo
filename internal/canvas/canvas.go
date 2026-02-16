package canvas

import (
	"encoding/json"
	"image/color"
	"os"
	"path/filepath"

	"gioui.org/f32"

	"screenpengo/internal/tool"
)

type Canvas struct {
	Strokes      []Stroke
	Current      *Stroke
	Shapes       []Shape
	CurrentShape *Shape
}

type Shape struct {
	Type     tool.ShapeType
	Color    color.NRGBA
	StartPos f32.Point
	EndPos   f32.Point
	WidthPx  float32
}

func (c *Canvas) StartStroke(color color.NRGBA, widthPx float32, startPoint f32.Point) {
	c.Current = &Stroke{
		Color:  color,
		Width:  widthPx,
		Points: []f32.Point{startPoint},
	}
}

func (c *Canvas) AddPoint(point f32.Point) {
	if c.Current == nil {
		return
	}
	last := c.Current.Points[len(c.Current.Points)-1]
	appendInterpolated(&c.Current.Points, last, point, c.Current.Width/2)
}

func (c *Canvas) FinishStroke() {
	if c.Current != nil {
		c.Strokes = append(c.Strokes, *c.Current)
		c.Current = nil
	}
}

func (c *Canvas) Clear() {
	c.Strokes = nil
	c.Current = nil
	c.Shapes = nil
	c.CurrentShape = nil
}

func (c *Canvas) StartShape(shapeType tool.ShapeType, color color.NRGBA, widthPx float32, startPoint f32.Point) {
	c.CurrentShape = &Shape{
		Type:     shapeType,
		Color:    color,
		StartPos: startPoint,
		EndPos:   startPoint,
		WidthPx:  widthPx,
	}
}

func (c *Canvas) UpdateShape(endPoint f32.Point) {
	if c.CurrentShape != nil {
		c.CurrentShape.EndPos = endPoint
	}
}

func (c *Canvas) FinishShape() {
	if c.CurrentShape != nil {
		c.Shapes = append(c.Shapes, *c.CurrentShape)
		c.CurrentShape = nil
	}
}

func (c *Canvas) RemoveShapesIntersectingStroke(stroke *Stroke) {
	if stroke == nil || len(stroke.Points) == 0 {
		return
	}

	remainingShapes := make([]Shape, 0, len(c.Shapes))
	for i := range c.Shapes {
		if !shapeIntersectsStroke(&c.Shapes[i], stroke) {
			remainingShapes = append(remainingShapes, c.Shapes[i])
		}
	}
	c.Shapes = remainingShapes
}

func shapeIntersectsStroke(shape *Shape, stroke *Stroke) bool {
	eraserRadius := stroke.Width * 2

	for _, point := range stroke.Points {
		if pointNearShape(point, shape, eraserRadius) {
			return true
		}
	}
	return false
}

func pointNearShape(point f32.Point, shape *Shape, radius float32) bool {
	minX := min(shape.StartPos.X, shape.EndPos.X) - radius
	maxX := max(shape.StartPos.X, shape.EndPos.X) + radius
	minY := min(shape.StartPos.Y, shape.EndPos.Y) - radius
	maxY := max(shape.StartPos.Y, shape.EndPos.Y) + radius

	if shape.Type == tool.Circle {
		dx := shape.EndPos.X - shape.StartPos.X
		dy := shape.EndPos.Y - shape.StartPos.Y
		circleRadius := float32(sqrtFloat64(float64(dx*dx + dy*dy)))

		minX = shape.StartPos.X - circleRadius - radius
		maxX = shape.StartPos.X + circleRadius + radius
		minY = shape.StartPos.Y - circleRadius - radius
		maxY = shape.StartPos.Y + circleRadius + radius
	}

	return point.X >= minX && point.X <= maxX &&
		point.Y >= minY && point.Y <= maxY
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func sqrtFloat64(x float64) float64 {
	if x < 0 {
		return 0
	}
	result := x
	for i := 0; i < 10; i++ {
		result = (result + x/result) / 2
	}
	return result
}

func (c *Canvas) SaveToFile(filename string) error {
	saveDir := filepath.Join(os.Getenv("HOME"), ".screenpen")
	if err := os.MkdirAll(saveDir, 0o755); err != nil {
		return err
	}

	fullPath := filepath.Join(saveDir, filename+".json")

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fullPath, data, 0o644)
}

func (c *Canvas) LoadFromFile(filename string) error {
	saveDir := filepath.Join(os.Getenv("HOME"), ".screenpen")
	fullPath := filepath.Join(saveDir, filename+".json")

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, c)
}

func ListSavedFiles() ([]string, error) {
	saveDir := filepath.Join(os.Getenv("HOME"), ".screenpen")

	if err := os.MkdirAll(saveDir, 0o755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(saveDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			name := entry.Name()
			files = append(files, name[:len(name)-5])
		}
	}

	return files, nil
}
