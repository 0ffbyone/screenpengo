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

// App coordinates all components and handles the frame update loop.
type App struct {
	canvas   *canvas.Canvas
	pen      *tool.PenConfig
	shape    *tool.ShapeConfig
	keyboard *input.KeyboardHandler
	pointer  *input.PointerHandler
	renderer *render.GioRenderer
	toolbar  *ui.Toolbar
	theme    *material.Theme

	cursorPos   f32.Point
	showCursor  bool
	isErasing   bool // Track if current stroke is an eraser stroke

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

// Frame processes input events and renders a frame.
func (a *App) Frame(gtx layout.Context) {
	// Pointer events should be scoped to the window rect.
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	event.Op(gtx.Ops, &a.ptrTag)
	area.Pop()

	// Check if any text input dialog is open
	dialogOpen := a.toolbar.IsDialogOpen()

	// Only capture keyboard focus if no dialog is open
	// This allows editors to receive keyboard input when dialogs are open
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

	// Render using Stack to layer canvas and toolbar
	layout.Stack{}.Layout(gtx,
		// Canvas layer (background)
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			// Calculate cursor radius in pixels
			cursorRadiusPixels := int(scaleToPixels(gtx, a.pen.WidthDp) / 2)
			cursorPosPixels := image.Point{
				X: int(a.cursorPos.X),
				Y: int(a.cursorPos.Y),
			}
			a.renderer.RenderFrame(gtx, a.canvas, cursorPosPixels, cursorRadiusPixels, a.showCursor)
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),
		// Toolbar layer (foreground)
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

			// Check if we're in eraser mode
			a.isErasing = (a.pen.ColorPreset == tool.Eraser)

			if a.shape.Active {
				// Start drawing a shape
				a.canvas.StartShape(a.shape.Type, a.pen.Color, widthInPixels, action.Position)
			} else {
				// Start drawing a freeform stroke
				a.canvas.StartStroke(a.pen.Color, widthInPixels, action.Position)
			}
			a.showCursor = false // Hide cursor while drawing
		case input.AddPoint:
			if a.shape.Active {
				// Update shape end position
				a.canvas.UpdateShape(action.Position)
			} else {
				// Add point to freeform stroke
				a.canvas.AddPoint(action.Position)
			}
		case input.FinishStroke:
			if a.shape.Active {
				// Finish the shape
				a.canvas.FinishShape()
			} else {
				// Finish the freeform stroke
				a.canvas.FinishStroke()

				// If we were erasing, remove shapes that intersect with the eraser stroke
				if a.isErasing && len(a.canvas.Strokes) > 0 {
					lastStroke := &a.canvas.Strokes[len(a.canvas.Strokes)-1]
					a.canvas.RemoveShapesIntersectingStroke(lastStroke)
				}
			}
			a.showCursor = true // Show cursor again after drawing
			a.isErasing = false
		case input.MoveCursor:
			a.cursorPos = action.Position
			a.showCursor = true
			gtx.Execute(op.InvalidateCmd{}) // Redraw to update cursor position
		}
	}

	// Keep animating while drawing.
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
	// Get current color, width, and state from toolbar
	color, widthDp, eraserClicked, slidersChanged, selectedShape, saveRequested, saveFilename, loadRequested, loadFilename := a.toolbar.HandleEvents(gtx)

	// Handle save request
	if saveRequested {
		if err := a.canvas.SaveToFile(saveFilename); err != nil {
			// TODO: Show error in UI
			println("Error saving file:", err.Error())
		} else {
			println("Saved to ~/.screenpen/" + saveFilename + ".json")
		}
	}

	// Handle load request
	if loadRequested {
		if err := a.canvas.LoadFromFile(loadFilename); err != nil {
			// TODO: Show error in UI
			println("Error loading file:", err.Error())
		} else {
			println("Loaded from ~/.screenpen/" + loadFilename + ".json")
			gtx.Execute(op.InvalidateCmd{}) // Redraw to show loaded content
		}
	}

	// Handle shape selection
	if selectedShape != tool.NoShape {
		a.shape.Type = selectedShape
		a.shape.Active = true
		// Deactivate eraser when shape is selected
		a.pen.ColorPreset = tool.Red
	}

	// If eraser was clicked, switch to eraser and deactivate shapes
	if eraserClicked {
		a.pen.SetColor(tool.Eraser)
		a.shape.Active = false
	} else if slidersChanged {
		// If sliders changed, update from slider values (exits eraser mode)
		a.pen.Color = color
		a.pen.WidthDp = widthDp
		// Clear the eraser preset when switching to custom color
		a.pen.ColorPreset = tool.Red // Default to something, actual color is from sliders
		// Deactivate shapes when drawing with sliders
		a.shape.Active = false
	}

	// Always invalidate to keep UI responsive
	gtx.Execute(op.InvalidateCmd{})
}

// scaleToPixels converts device-independent units to physical screen pixels
// using the display's pixel density.
func scaleToPixels(gtx layout.Context, deviceIndependentValue float32) float32 {
	return float32(gtx.Metric.PxPerDp) * deviceIndependentValue
}
