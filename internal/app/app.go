package app

import (
	"image/color"
	"os"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget/material"

	"screenpengo/internal/canvas"
	"screenpengo/internal/input"
	"screenpengo/internal/render"
	"screenpengo/internal/tool"
	"screenpengo/internal/ui"
)

// App coordinates all components and handles the frame update loop.
type App struct {
	canvas   *canvas.Canvas
	pen      *tool.PenConfig
	keyboard *input.KeyboardHandler
	pointer  *input.PointerHandler
	renderer *render.GioRenderer
	toolbar  *ui.Toolbar
	theme    *material.Theme

	keyTag struct{}
	ptrTag struct{}
}

// New creates a new App instance with default settings.
func New() *App {
	theme := material.NewTheme()
	return &App{
		canvas: &canvas.Canvas{},
		pen: &tool.PenConfig{
			Color:       color.NRGBA{R: 255, A: 255}, // red default
			WidthDp:     6,
			ColorPreset: tool.Red,
			WidthPreset: tool.Medium,
		},
		keyboard: input.NewKeyboardHandler(),
		pointer:  input.NewPointerHandler(),
		renderer: &render.GioRenderer{},
		toolbar:  ui.NewToolbar(theme),
		theme:    theme,
	}
}

// Frame processes input events and renders a frame.
func (a *App) Frame(gtx layout.Context) {
	// Pointer events should be scoped to the window rect.
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	event.Op(gtx.Ops, &a.ptrTag)
	area.Pop()

	// Keyboard focus/events.
	event.Op(gtx.Ops, &a.keyTag)
	key.InputHintOp{Tag: &a.keyTag, Hint: key.HintAny}.Add(gtx.Ops)
	gtx.Execute(key.FocusCmd{Tag: &a.keyTag})

	a.applyPointerActions(gtx)
	a.applyKeyboardActions(gtx)
	a.applyToolbarActions(gtx)

	// Render using Stack to layer canvas and toolbar
	layout.Stack{}.Layout(gtx,
		// Canvas layer (background)
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			a.renderer.RenderFrame(gtx, a.canvas)
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),
		// Toolbar layer (foreground)
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return a.toolbar.Layout(gtx, a.pen.ColorPreset, a.pen.WidthPreset)
		}),
	)
}

func (a *App) applyPointerActions(gtx layout.Context) {
	actions := a.pointer.HandleEvents(gtx, &a.ptrTag)

	for _, action := range actions {
		switch action.Type {
		case input.StartStroke:
			widthInPixels := scaleToPixels(gtx, a.pen.WidthDp)
			a.canvas.StartStroke(a.pen.Color, widthInPixels, action.Position)
		case input.AddPoint:
			a.canvas.AddPoint(action.Position)
		case input.FinishStroke:
			a.canvas.FinishStroke()
		}
	}

	// Keep animating while drawing.
	if a.canvas.Current != nil {
		gtx.Execute(op.InvalidateCmd{})
	}
}

func (a *App) applyKeyboardActions(gtx layout.Context) {
	actions := a.keyboard.HandleEvents(gtx, &a.keyTag)

	for _, action := range actions {
		switch action.Type {
		case input.SetColor:
			a.pen.SetColor(action.ColorPreset)
		case input.SetWidth:
			a.pen.SetWidth(action.WidthPreset)
		case input.ToggleDim:
			a.renderer.Dim = !a.renderer.Dim
		case input.Clear:
			a.canvas.Clear()
		case input.Quit:
			os.Exit(0)
		}
	}

	if len(actions) > 0 {
		gtx.Execute(op.InvalidateCmd{})
	}
}

func (a *App) applyToolbarActions(gtx layout.Context) {
	actions := a.toolbar.HandleEvents(gtx)

	for _, action := range actions {
		switch action.Type {
		case input.SetColor:
			a.pen.SetColor(action.ColorPreset)
		case input.SetWidth:
			a.pen.SetWidth(action.WidthPreset)
		}
	}

	if len(actions) > 0 {
		gtx.Execute(op.InvalidateCmd{})
	}
}

// scaleToPixels converts device-independent units to physical screen pixels
// using the display's pixel density.
func scaleToPixels(gtx layout.Context, deviceIndependentValue float32) float32 {
	return float32(gtx.Metric.PxPerDp) * deviceIndependentValue
}
