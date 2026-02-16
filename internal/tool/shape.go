package tool

type ShapeType int

const (
	NoShape ShapeType = iota
	Circle
	Rectangle
	Line
	Arrow
)

type ShapeConfig struct {
	Type   ShapeType
	Active bool
}
