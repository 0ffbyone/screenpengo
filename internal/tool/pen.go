package tool

import "image/color"

// ColorPreset represents a predefined pen color.
type ColorPreset int

const (
	Red ColorPreset = iota
	Green
	Blue
	Yellow
	Orange
	Pink
	Blur
	Eraser
)

// WidthPreset represents a predefined pen width level.
type WidthPreset int

const (
	Thin WidthPreset = iota
	Medium
	Thick
)

// PenConfig holds the current pen color and width settings.
type PenConfig struct {
	Color       color.NRGBA
	WidthDp     float32
	ColorPreset ColorPreset
	WidthPreset WidthPreset
}

// SetColor sets the pen color to a predefined preset.
func (p *PenConfig) SetColor(preset ColorPreset) {
	p.ColorPreset = preset
	switch preset {
	case Red:
		p.Color = color.NRGBA{R: 255, A: 255}
	case Green:
		p.Color = color.NRGBA{G: 255, A: 255}
	case Blue:
		p.Color = color.NRGBA{B: 255, A: 255}
	case Yellow:
		p.Color = color.NRGBA{R: 255, G: 255, A: 255}
	case Orange:
		p.Color = color.NRGBA{R: 255, G: 165, A: 255}
	case Pink:
		p.Color = color.NRGBA{R: 255, G: 105, B: 180, A: 255}
	case Blur:
		// "Blur" pen: wide semi-transparent black.
		p.Color = color.NRGBA{A: 0x40}
		p.WidthDp = 20
		p.WidthPreset = Thick
	case Eraser:
		// Eraser: opaque white, keeps current width
		p.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		// Don't change width - keep whatever width is currently set
	}
}

// SetWidth sets the pen width to a predefined level.
func (p *PenConfig) SetWidth(preset WidthPreset) {
	p.WidthPreset = preset
	switch preset {
	case Thin:
		p.WidthDp = 3
	case Medium:
		p.WidthDp = 6
	case Thick:
		p.WidthDp = 12
	}
}
