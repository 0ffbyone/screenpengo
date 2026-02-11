package tool

// ShapeType represents different geometric shapes that can be drawn.
type ShapeType int

const (
	NoShape ShapeType = iota
	Circle
	Rectangle
	Line
	Arrow
)

// ShapeConfig holds the current shape drawing settings.
type ShapeConfig struct {
	Type   ShapeType
	Active bool
}
