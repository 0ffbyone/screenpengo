package tool

import "image/color"

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

type WidthPreset int

const (
	Thin WidthPreset = iota
	Medium
	Thick
)

type PenConfig struct {
	Color       color.NRGBA
	WidthDp     float32
	ColorPreset ColorPreset
	WidthPreset WidthPreset
}

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
		p.Color = color.NRGBA{A: 0x40}
		p.WidthDp = 20
		p.WidthPreset = Thick
	case Eraser:
		p.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}
}

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
