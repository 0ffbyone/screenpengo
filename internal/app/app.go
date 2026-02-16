package app

import (
	"image"
	"image/color"
	"os"

	"gioui.org/f32"
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

type App struct {
	canvas   *canvas.Canvas
	pen      *tool.PenConfig
	shape    *tool.ShapeConfig
	keyboard *input.KeyboardHandler
	pointer  *input.PointerHandler
	renderer *render.GioRenderer
	toolbar  *ui.Toolbar
	theme    *material.Theme

	cursorPos  f32.Point
	showCursor bool
	isErasing  bool

	keyTag struct{}
	ptrTag struct{}
}

func New() *App {
	theme := material.NewTheme()
	return &App{
		canvas: &canvas.Canvas{},
		pen: &tool.PenConfig{
			Color:       color.NRGBA{R: 255, A: 255},
			WidthDp:     6,
			ColorPreset: tool.Red,
			WidthPreset: tool.Medium,
		},
		shape: &tool.ShapeConfig{
			Type:   tool.NoShape,
			Active: false,
		},
		keyboard: input.NewKeyboardHandler(),
		pointer:  input.NewPointerHandler(),
		renderer: &render.GioRenderer{},
		toolbar:  ui.NewToolbar(theme),
		theme:    theme,
	}
}

func (a *App) Frame(gtx layout.Context) {
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	event.Op(gtx.Ops, &a.ptrTag)
	area.Pop()

	dialogOpen := a.toolbar.IsDialogOpen()

	if !dialogOpen {
		event.Op(gtx.Ops, &a.keyTag)
		key.InputHintOp{Tag: &a.keyTag, Hint: key.HintAny}.Add(gtx.Ops)
		gtx.Execute(key.FocusCmd{Tag: &a.keyTag})
	}

	a.applyPointerActions(gtx)
	if !dialogOpen {
		a.applyKeyboardActions(gtx)
	}
	a.applyToolbarActions(gtx)

	layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			cursorRadiusPixels := int(scaleToPixels(gtx, a.pen.WidthDp) / 2)
			cursorPosPixels := image.Point{
				X: int(a.cursorPos.X),
				Y: int(a.cursorPos.Y),
			}
			a.renderer.RenderFrame(gtx, a.canvas, cursorPosPixels, cursorRadiusPixels, a.showCursor)
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return a.toolbar.Layout(gtx)
		}),
	)
}

func (a *App) applyPointerActions(gtx layout.Context) {
	actions := a.pointer.HandleEvents(gtx, &a.ptrTag)

	for _, action := range actions {
		switch action.Type {
		case input.StartStroke:
			widthInPixels := scaleToPixels(gtx, a.pen.WidthDp)

			a.isErasing = (a.pen.ColorPreset == tool.Eraser)

			if a.shape.Active {
				a.canvas.StartShape(a.shape.Type, a.pen.Color, widthInPixels, action.Position)
			} else {
				a.canvas.StartStroke(a.pen.Color, widthInPixels, action.Position)
			}
			a.showCursor = false
		case input.AddPoint:
			if a.shape.Active {
				a.canvas.UpdateShape(action.Position)
			} else {
				a.canvas.AddPoint(action.Position)
			}
		case input.FinishStroke:
			if a.shape.Active {
				a.canvas.FinishShape()
			} else {
				a.canvas.FinishStroke()

				if a.isErasing && len(a.canvas.Strokes) > 0 {
					lastStroke := &a.canvas.Strokes[len(a.canvas.Strokes)-1]
					a.canvas.RemoveShapesIntersectingStroke(lastStroke)
				}
			}
			a.showCursor = true
			a.isErasing = false
		case input.MoveCursor:
			a.cursorPos = action.Position
			a.showCursor = true
			gtx.Execute(op.InvalidateCmd{})
		}
	}

	if a.canvas.Current != nil || a.canvas.CurrentShape != nil {
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
	color, widthDp, eraserClicked, slidersChanged, selectedShape, saveRequested, saveFilename, loadRequested, loadFilename := a.toolbar.HandleEvents(gtx)

	if saveRequested {
		if err := a.canvas.SaveToFile(saveFilename); err != nil {
			println("Error saving file:", err.Error())
		} else {
			println("Saved to ~/.screenpen/" + saveFilename + ".json")
		}
	}

	if loadRequested {
		if err := a.canvas.LoadFromFile(loadFilename); err != nil {
			println("Error loading file:", err.Error())
		} else {
			println("Loaded from ~/.screenpen/" + loadFilename + ".json")
			gtx.Execute(op.InvalidateCmd{})
		}
	}

	if selectedShape != tool.NoShape {
		a.shape.Type = selectedShape
		a.shape.Active = true
		a.pen.ColorPreset = tool.Red
	}

	if eraserClicked {
		a.pen.SetColor(tool.Eraser)
		a.shape.Active = false
	} else if slidersChanged {
		a.pen.Color = color
		a.pen.WidthDp = widthDp
		a.pen.ColorPreset = tool.Red
		a.shape.Active = false
	}

	gtx.Execute(op.InvalidateCmd{})
}

func scaleToPixels(gtx layout.Context, deviceIndependentValue float32) float32 {
	return float32(gtx.Metric.PxPerDp) * deviceIndependentValue
}
