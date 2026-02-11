package canvas

import (
	"encoding/json"
	"image/color"
	"os"
	"path/filepath"

	"gioui.org/f32"

	"screenpengo/internal/tool"
)

// Canvas manages all drawing strokes and shapes.
type Canvas struct {
	Strokes      []Stroke
	Current      *Stroke
	Shapes       []Shape
	CurrentShape *Shape
}

// Shape represents a geometric shape drawn on the canvas.
type Shape struct {
	Type      tool.ShapeType
	Color     color.NRGBA
	StartPos  f32.Point
	EndPos    f32.Point
	WidthPx   float32
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

// Clear removes all strokes and shapes from the canvas.
func (c *Canvas) Clear() {
	c.Strokes = nil
	c.Current = nil
	c.Shapes = nil
	c.CurrentShape = nil
}

// StartShape begins a new shape with the given type, color, and start position.
func (c *Canvas) StartShape(shapeType tool.ShapeType, color color.NRGBA, widthPx float32, startPoint f32.Point) {
	c.CurrentShape = &Shape{
		Type:     shapeType,
		Color:    color,
		StartPos: startPoint,
		EndPos:   startPoint,
		WidthPx:  widthPx,
	}
}

// UpdateShape updates the end position of the current shape.
func (c *Canvas) UpdateShape(endPoint f32.Point) {
	if c.CurrentShape != nil {
		c.CurrentShape.EndPos = endPoint
	}
}

// FinishShape commits the current shape to the canvas.
func (c *Canvas) FinishShape() {
	if c.CurrentShape != nil {
		c.Shapes = append(c.Shapes, *c.CurrentShape)
		c.CurrentShape = nil
	}
}

// RemoveShapesIntersectingStroke removes any shapes that intersect with the given stroke (for eraser).
func (c *Canvas) RemoveShapesIntersectingStroke(stroke *Stroke) {
	if stroke == nil || len(stroke.Points) == 0 {
		return
	}

	// Filter out shapes that intersect with the eraser stroke
	remainingShapes := make([]Shape, 0, len(c.Shapes))
	for i := range c.Shapes {
		if !shapeIntersectsStroke(&c.Shapes[i], stroke) {
			remainingShapes = append(remainingShapes, c.Shapes[i])
		}
	}
	c.Shapes = remainingShapes
}

// shapeIntersectsStroke checks if any point in the stroke is near the shape
func shapeIntersectsStroke(shape *Shape, stroke *Stroke) bool {
	eraserRadius := stroke.Width * 2 // Generous hit detection

	for _, point := range stroke.Points {
		if pointNearShape(point, shape, eraserRadius) {
			return true
		}
	}
	return false
}

// pointNearShape checks if a point is close to the shape
func pointNearShape(point f32.Point, shape *Shape, radius float32) bool {
	// For simplicity, check if point is near the bounding box of the shape
	// This works reasonably well for all shape types

	minX := min(shape.StartPos.X, shape.EndPos.X) - radius
	maxX := max(shape.StartPos.X, shape.EndPos.X) + radius
	minY := min(shape.StartPos.Y, shape.EndPos.Y) - radius
	maxY := max(shape.StartPos.Y, shape.EndPos.Y) + radius

	// For circles, expand the bounding box by the radius
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

// Helper functions
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
	// Simple approximation - could use math.Sqrt but avoiding import
	result := x
	for i := 0; i < 10; i++ {
		result = (result + x/result) / 2
	}
	return result
}

// SaveToFile saves the canvas state to a JSON file
func (c *Canvas) SaveToFile(filename string) error {
	// Create save directory if it doesn't exist
	saveDir := filepath.Join(os.Getenv("HOME"), ".screenpen")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return err
	}

	// Full path
	fullPath := filepath.Join(saveDir, filename+".json")

	// Serialize canvas to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(fullPath, data, 0644)
}

// LoadFromFile loads the canvas state from a JSON file
func (c *Canvas) LoadFromFile(filename string) error {
	// Full path
	saveDir := filepath.Join(os.Getenv("HOME"), ".screenpen")
	fullPath := filepath.Join(saveDir, filename+".json")

	// Read file
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	// Deserialize JSON to canvas
	return json.Unmarshal(data, c)
}

// ListSavedFiles returns a list of saved drawing filenames (without .json extension)
func ListSavedFiles() ([]string, error) {
	saveDir := filepath.Join(os.Getenv("HOME"), ".screenpen")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, err
	}

	// Read directory
	entries, err := os.ReadDir(saveDir)
	if err != nil {
		return nil, err
	}

	// Filter for .json files and remove extension
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			name := entry.Name()
			files = append(files, name[:len(name)-5]) // Remove ".json"
		}
	}

	return files, nil
}
