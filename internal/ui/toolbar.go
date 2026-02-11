package ui

import (
	"fmt"
	"image/color"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"screenpengo/internal/input"
	"screenpengo/internal/tool"
)

// Button label constants
const (
	labelRed    = "R"
	labelGreen  = "G"
	labelBlue   = "B"
	labelYellow = "Y"
	labelOrange = "O"
	labelPink   = "P"
	labelBlur   = "X"
	labelThin   = "Thin"
	labelMedium = "Med"
	labelThick  = "Thick"
)

// Toolbar provides a UI overlay for selecting colors and widths.
type Toolbar struct {
	// Color buttons
	btnRed    widget.Clickable
	btnGreen  widget.Clickable
	btnBlue   widget.Clickable
	btnYellow widget.Clickable
	btnOrange widget.Clickable
	btnPink   widget.Clickable
	btnBlur   widget.Clickable

	// Width buttons
	btnThin   widget.Clickable
	btnMedium widget.Clickable
	btnThick  widget.Clickable

	// For event filtering
	panelTag struct{}

	theme *material.Theme
}

// NewToolbar creates a new toolbar with the given theme.
func NewToolbar(theme *material.Theme) *Toolbar {
	return &Toolbar{theme: theme}
}

// HandleEvents processes toolbar button clicks and returns actions.
func (t *Toolbar) HandleEvents(gtx layout.Context) []input.Action {
	var actions []input.Action

	// Color buttons
	if t.btnRed.Clicked(gtx) {
		actions = append(actions, input.Action{
			Type:        input.SetColor,
			ColorPreset: tool.Red,
		})
	}
	if t.btnGreen.Clicked(gtx) {
		actions = append(actions, input.Action{
			Type:        input.SetColor,
			ColorPreset: tool.Green,
		})
	}
	if t.btnBlue.Clicked(gtx) {
		actions = append(actions, input.Action{
			Type:        input.SetColor,
			ColorPreset: tool.Blue,
		})
	}
	if t.btnYellow.Clicked(gtx) {
		actions = append(actions, input.Action{
			Type:        input.SetColor,
			ColorPreset: tool.Yellow,
		})
	}
	if t.btnOrange.Clicked(gtx) {
		actions = append(actions, input.Action{
			Type:        input.SetColor,
			ColorPreset: tool.Orange,
		})
	}
	if t.btnPink.Clicked(gtx) {
		actions = append(actions, input.Action{
			Type:        input.SetColor,
			ColorPreset: tool.Pink,
		})
	}
	if t.btnBlur.Clicked(gtx) {
		actions = append(actions, input.Action{
			Type:        input.SetColor,
			ColorPreset: tool.Blur,
		})
	}

	// Width buttons
	if t.btnThin.Clicked(gtx) {
		actions = append(actions, input.Action{
			Type:        input.SetWidth,
			WidthPreset: tool.Thin,
		})
	}
	if t.btnMedium.Clicked(gtx) {
		actions = append(actions, input.Action{
			Type:        input.SetWidth,
			WidthPreset: tool.Medium,
		})
	}
	if t.btnThick.Clicked(gtx) {
		actions = append(actions, input.Action{
			Type:        input.SetWidth,
			WidthPreset: tool.Thick,
		})
	}

	return actions
}

// Layout renders the toolbar as a left sidebar.
func (t *Toolbar) Layout(gtx layout.Context, currentColor tool.ColorPreset, currentWidth tool.WidthPreset) layout.Dimensions {
	fmt.Printf("Toolbar.Layout called - Constraints: Min=%v Max=%v\n", gtx.Constraints.Min, gtx.Constraints.Max)
	dims := t.layoutPanel(gtx, currentColor, currentWidth)
	fmt.Printf("Toolbar dimensions: %v\n", dims.Size)
	return dims
}

func (t *Toolbar) layoutPanel(gtx layout.Context, currentColor tool.ColorPreset, currentWidth tool.WidthPreset) layout.Dimensions {
	return layout.Inset{Top: 20, Bottom: 20, Left: 10, Right: 10}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		backgroundColor := color.NRGBA{R: 255, G: 255, B: 255, A: 240}

		// Use Stack to layer background behind buttons
		return layout.Stack{}.Layout(gtx,
			// Background layer
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				// Draw background
				fmt.Printf("Drawing background sized to: %v\n", gtx.Constraints.Min)
				defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()
				paint.ColorOp{Color: backgroundColor}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				return layout.Dimensions{Size: gtx.Constraints.Min}
			}),
			// Button layer on top
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: 10, Bottom: 10, Left: 10, Right: 10}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					// Capture events
					defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
					event.Op(gtx.Ops, &t.panelTag)
					for {
						ev, ok := gtx.Event(pointer.Filter{
							Target: &t.panelTag,
							Kinds:  pointer.Press | pointer.Drag | pointer.Release,
						})
						if !ok {
							break
						}
						_ = ev
					}

					// Buttons
					return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceBetween}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return t.layoutColorButtons(gtx, currentColor)
						}),
						layout.Rigid(layout.Spacer{Height: 20}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return t.layoutWidthButtons(gtx, currentWidth)
						}),
					)
				})
			}),
		)
	})
}

func (t *Toolbar) layoutColorButtons(gtx layout.Context, current tool.ColorPreset) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
		layout.Rigid(t.makeColorButton(labelRed, tool.Red, &t.btnRed, current == tool.Red)),
		layout.Rigid(layout.Spacer{Height: 5}.Layout),
		layout.Rigid(t.makeColorButton(labelGreen, tool.Green, &t.btnGreen, current == tool.Green)),
		layout.Rigid(layout.Spacer{Height: 5}.Layout),
		layout.Rigid(t.makeColorButton(labelBlue, tool.Blue, &t.btnBlue, current == tool.Blue)),
		layout.Rigid(layout.Spacer{Height: 5}.Layout),
		layout.Rigid(t.makeColorButton(labelYellow, tool.Yellow, &t.btnYellow, current == tool.Yellow)),
		layout.Rigid(layout.Spacer{Height: 5}.Layout),
		layout.Rigid(t.makeColorButton(labelOrange, tool.Orange, &t.btnOrange, current == tool.Orange)),
		layout.Rigid(layout.Spacer{Height: 5}.Layout),
		layout.Rigid(t.makeColorButton(labelPink, tool.Pink, &t.btnPink, current == tool.Pink)),
		layout.Rigid(layout.Spacer{Height: 5}.Layout),
		layout.Rigid(t.makeColorButton(labelBlur, tool.Blur, &t.btnBlur, current == tool.Blur)),
	)
}

func (t *Toolbar) layoutWidthButtons(gtx layout.Context, current tool.WidthPreset) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
		layout.Rigid(t.makeWidthButton(labelThin, tool.Thin, &t.btnThin, current == tool.Thin)),
		layout.Rigid(layout.Spacer{Height: 5}.Layout),
		layout.Rigid(t.makeWidthButton(labelMedium, tool.Medium, &t.btnMedium, current == tool.Medium)),
		layout.Rigid(layout.Spacer{Height: 5}.Layout),
		layout.Rigid(t.makeWidthButton(labelThick, tool.Thick, &t.btnThick, current == tool.Thick)),
	)
}

func (t *Toolbar) makeColorButton(label string, _ tool.ColorPreset, clk *widget.Clickable, selected bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		btn := material.Button(t.theme, clk, label)
		btn.CornerRadius = 8

		if selected {
			btn.Background = color.NRGBA{R: 70, G: 130, B: 220, A: 255}
		} else {
			btn.Background = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
		}

		gtx.Constraints.Min.X = gtx.Dp(60)
		return btn.Layout(gtx)
	}
}

func (t *Toolbar) makeWidthButton(label string, _ tool.WidthPreset, clk *widget.Clickable, selected bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		btn := material.Button(t.theme, clk, label)
		btn.CornerRadius = 8

		if selected {
			btn.Background = color.NRGBA{R: 70, G: 130, B: 220, A: 255}
		} else {
			btn.Background = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
		}

		gtx.Constraints.Min.X = gtx.Dp(60)
		return btn.Layout(gtx)
	}
}
