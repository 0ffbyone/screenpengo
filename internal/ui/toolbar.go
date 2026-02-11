package ui

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"screenpengo/internal/canvas"
	"screenpengo/internal/tool"
)

// Toolbar provides a UI overlay for selecting colors and widths.
type Toolbar struct {
	// Main buttons
	colorButton  widget.Clickable
	widthButton  widget.Clickable
	eraserButton widget.Clickable
	shapesButton widget.Clickable
	saveButton   widget.Clickable
	loadButton   widget.Clickable

	// Shape buttons
	circleButton    widget.Clickable
	rectangleButton widget.Clickable
	lineButton      widget.Clickable
	arrowButton     widget.Clickable

	// Save panel buttons
	confirmSaveButton widget.Clickable
	cancelSaveButton  widget.Clickable

	// Load panel buttons
	confirmLoadButton widget.Clickable
	cancelLoadButton  widget.Clickable

	// Text input for filename (used by both save and load)
	filenameEditor     widget.Editor
	loadFilenameEditor widget.Editor

	// File list for load dialog
	savedFiles       []string
	fileButtons      []widget.Clickable
	refreshListButton widget.Clickable

	// Color sliders (shown when color picker is open)
	redSlider   widget.Float
	greenSlider widget.Float
	blueSlider  widget.Float

	// Width slider (shown when width picker is open)
	widthSlider widget.Float

	// Previous slider values to detect changes
	prevRedValue   float32
	prevGreenValue float32
	prevBlueValue  float32
	prevWidthValue float32

	// State to track which panel is open
	colorPickerOpen  bool
	widthPickerOpen  bool
	shapesPickerOpen bool
	saveDialogOpen   bool
	loadDialogOpen   bool

	// Track if eraser is currently active
	eraserActive bool

	theme *material.Theme
}

// NewToolbar creates a new toolbar with the given theme.
func NewToolbar(theme *material.Theme) *Toolbar {
	saveEditor := widget.Editor{
		SingleLine: true,
		Submit:     true,
	}
	saveEditor.SetText("") // Start empty so user enters custom name

	loadEditor := widget.Editor{
		SingleLine: true,
		Submit:     true,
	}
	loadEditor.SetText("")

	return &Toolbar{
		theme: theme,
		// Initialize sliders with default values
		redSlider:          widget.Float{Value: 1.0}, // Red default
		greenSlider:        widget.Float{Value: 0.0},
		blueSlider:         widget.Float{Value: 0.0},
		widthSlider:        widget.Float{Value: 0.5}, // Medium width default
		filenameEditor:     saveEditor,
		loadFilenameEditor: loadEditor,
	}
}

// HandleEvents processes toolbar button clicks and slider changes, returns the current color, width, and state flags.
func (t *Toolbar) HandleEvents(gtx layout.Context) (currentColor color.NRGBA, currentWidth float32, eraserClicked bool, slidersChanged bool, selectedShape tool.ShapeType, saveRequested bool, saveFilename string, loadRequested bool, loadFilename string) {
	selectedShape = tool.NoShape
	saveRequested = false
	loadRequested = false
	// Handle button clicks
	if t.colorButton.Clicked(gtx) {
		t.colorPickerOpen = !t.colorPickerOpen
		// Close width picker when opening color picker
		if t.colorPickerOpen {
			t.widthPickerOpen = false
		}
	}

	if t.widthButton.Clicked(gtx) {
		t.widthPickerOpen = !t.widthPickerOpen
		// Close other pickers when opening width picker
		if t.widthPickerOpen {
			t.colorPickerOpen = false
			t.shapesPickerOpen = false
		}
	}

	if t.shapesButton.Clicked(gtx) {
		t.shapesPickerOpen = !t.shapesPickerOpen
		// Close other pickers when opening shapes picker
		if t.shapesPickerOpen {
			t.colorPickerOpen = false
			t.widthPickerOpen = false
		}
	}

	// Handle shape selection
	if t.circleButton.Clicked(gtx) {
		selectedShape = tool.Circle
		t.eraserActive = false
	}
	if t.rectangleButton.Clicked(gtx) {
		selectedShape = tool.Rectangle
		t.eraserActive = false
	}
	if t.lineButton.Clicked(gtx) {
		selectedShape = tool.Line
		t.eraserActive = false
	}
	if t.arrowButton.Clicked(gtx) {
		selectedShape = tool.Arrow
		t.eraserActive = false
	}

	// Handle save button
	if t.saveButton.Clicked(gtx) {
		t.saveDialogOpen = !t.saveDialogOpen
		// Close other pickers when opening save dialog
		if t.saveDialogOpen {
			t.colorPickerOpen = false
			t.widthPickerOpen = false
			t.shapesPickerOpen = false
			t.loadDialogOpen = false
		}
	}

	// Handle save dialog actions
	if t.confirmSaveButton.Clicked(gtx) {
		saveRequested = true
		saveFilename = t.filenameEditor.Text()
		t.saveDialogOpen = false
	}
	if t.cancelSaveButton.Clicked(gtx) {
		t.saveDialogOpen = false
	}

	// Handle load button
	if t.loadButton.Clicked(gtx) {
		t.loadDialogOpen = !t.loadDialogOpen
		// Close other pickers when opening load dialog
		if t.loadDialogOpen {
			t.colorPickerOpen = false
			t.widthPickerOpen = false
			t.shapesPickerOpen = false
			t.saveDialogOpen = false
			// Refresh file list when opening load dialog
			t.refreshFileList()
		}
	}

	// Handle refresh list button
	if t.refreshListButton.Clicked(gtx) {
		t.refreshFileList()
	}

	// Handle file button clicks
	for i := range t.fileButtons {
		if t.fileButtons[i].Clicked(gtx) {
			loadRequested = true
			loadFilename = t.savedFiles[i]
			t.loadDialogOpen = false
			break
		}
	}

	// Handle load dialog actions
	if t.confirmLoadButton.Clicked(gtx) {
		loadRequested = true
		loadFilename = t.loadFilenameEditor.Text()
		t.loadDialogOpen = false
	}
	if t.cancelLoadButton.Clicked(gtx) {
		t.loadDialogOpen = false
	}

	if t.eraserButton.Clicked(gtx) {
		// Toggle eraser mode
		t.eraserActive = !t.eraserActive

		if t.eraserActive {
			// Entering eraser mode
			eraserClicked = true
			// Close all pickers when eraser is activated
			t.colorPickerOpen = false
			t.widthPickerOpen = false
		} else {
			// Exiting eraser mode - restore slider colors
			slidersChanged = true
		}
	}

	// Check if sliders have changed
	if t.redSlider.Value != t.prevRedValue ||
	   t.greenSlider.Value != t.prevGreenValue ||
	   t.blueSlider.Value != t.prevBlueValue ||
	   t.widthSlider.Value != t.prevWidthValue {
		slidersChanged = true
		t.eraserActive = false // Deactivate eraser when sliders change
		t.prevRedValue = t.redSlider.Value
		t.prevGreenValue = t.greenSlider.Value
		t.prevBlueValue = t.blueSlider.Value
		t.prevWidthValue = t.widthSlider.Value
	}

	// Get color from RGB sliders (0.0 to 1.0)
	currentColor = color.NRGBA{
		R: uint8(t.redSlider.Value * 255),
		G: uint8(t.greenSlider.Value * 255),
		B: uint8(t.blueSlider.Value * 255),
		A: 255,
	}

	// Get width from slider (map 0.0-1.0 to 2-20 dp)
	currentWidth = 2 + (t.widthSlider.Value * 18)

	return currentColor, currentWidth, eraserClicked, slidersChanged, selectedShape, saveRequested, saveFilename, loadRequested, loadFilename
}

// Layout renders the toolbar as a left sidebar, vertically centered.
func (t *Toolbar) Layout(gtx layout.Context) layout.Dimensions {
	// Vertically center the toolbar
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, layout.Spacer{}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Horizontal layout: buttons on left, picker panel on right (if open)
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				// Main toolbar buttons
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return t.layoutMainButtons(gtx)
				}),
				// Picker panel (color, width, shapes, save, or load dialog) - conditionally rendered
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if t.colorPickerOpen {
						return t.layoutColorPicker(gtx)
					} else if t.widthPickerOpen {
						return t.layoutWidthPicker(gtx)
					} else if t.shapesPickerOpen {
						return t.layoutShapesPicker(gtx)
					} else if t.saveDialogOpen {
						return t.layoutSaveDialog(gtx)
					} else if t.loadDialogOpen {
						return t.layoutLoadDialog(gtx)
					}
					return layout.Dimensions{}
				}),
			)
		}),
		layout.Flexed(1, layout.Spacer{}.Layout),
	)
}

// layoutMainButtons renders the compact vertical button bar
func (t *Toolbar) layoutMainButtons(gtx layout.Context) layout.Dimensions {
	return t.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Color button
				btn := material.Button(t.theme, &t.colorButton, "Color")
				btn.Background = color.NRGBA{R: 70, G: 70, B: 70, A: 220}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Width button
				btn := material.Button(t.theme, &t.widthButton, "Width")
				btn.Background = color.NRGBA{R: 70, G: 70, B: 70, A: 220}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Eraser button - highlight when active
				btn := material.Button(t.theme, &t.eraserButton, "Eraser")
				if t.eraserActive {
					// Bright highlight when eraser is active
					btn.Background = color.NRGBA{R: 100, G: 180, B: 255, A: 255}
				} else {
					// Light gray when inactive
					btn.Background = color.NRGBA{R: 220, G: 220, B: 220, A: 220}
				}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Shapes button
				btn := material.Button(t.theme, &t.shapesButton, "Shapes")
				btn.Background = color.NRGBA{R: 70, G: 70, B: 70, A: 220}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Save button
				btn := material.Button(t.theme, &t.saveButton, "Save")
				btn.Background = color.NRGBA{R: 50, G: 150, B: 50, A: 220}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Load button
				btn := material.Button(t.theme, &t.loadButton, "Load")
				btn.Background = color.NRGBA{R: 50, G: 100, B: 200, A: 220}
				return btn.Layout(gtx)
			}),
		)
	})
}

// layoutColorPicker renders the RGB sliders panel
func (t *Toolbar) layoutColorPicker(gtx layout.Context) layout.Dimensions {
	return t.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return t.layoutColorSliders(gtx)
	})
}

// layoutWidthPicker renders the width slider panel
func (t *Toolbar) layoutWidthPicker(gtx layout.Context) layout.Dimensions {
	return t.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return t.layoutWidthSlider(gtx)
	})
}

// layoutShapesPicker renders the shapes selection panel
func (t *Toolbar) layoutShapesPicker(gtx layout.Context) layout.Dimensions {
	return t.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(t.theme, &t.circleButton, "Circle")
				btn.Background = color.NRGBA{R: 80, G: 120, B: 180, A: 220}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(t.theme, &t.rectangleButton, "Rectangle")
				btn.Background = color.NRGBA{R: 80, G: 120, B: 180, A: 220}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(t.theme, &t.lineButton, "Line")
				btn.Background = color.NRGBA{R: 80, G: 120, B: 180, A: 220}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(t.theme, &t.arrowButton, "Arrow")
				btn.Background = color.NRGBA{R: 80, G: 120, B: 180, A: 220}
				return btn.Layout(gtx)
			}),
		)
	})
}

// layoutSaveDialog renders the save dialog with filename input
func (t *Toolbar) layoutSaveDialog(gtx layout.Context) layout.Dimensions {
	return t.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
			// Title
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(t.theme, "Save Drawing")
				label.Font.Weight = 700 // Bold
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			// Filename label
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body2(t.theme, "Filename:")
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			// Filename input
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Set fixed width for editor
				gtx.Constraints.Min.X = gtx.Dp(200)
				gtx.Constraints.Max.X = gtx.Dp(200)

				editor := material.Editor(t.theme, &t.filenameEditor, "Enter filename...")
				editor.TextSize = 14
				return editor.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			// Info text
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Caption(t.theme, "Saved to ~/.screenpen/")
				label.Color = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 15}.Layout),
			// Buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceStart}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(t.theme, &t.confirmSaveButton, "Save")
						btn.Background = color.NRGBA{R: 50, G: 150, B: 50, A: 255}
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 10}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(t.theme, &t.cancelSaveButton, "Cancel")
						btn.Background = color.NRGBA{R: 150, G: 50, B: 50, A: 255}
						return btn.Layout(gtx)
					}),
				)
			}),
		)
	})
}

// layoutLoadDialog renders the load dialog with list of saved files
func (t *Toolbar) layoutLoadDialog(gtx layout.Context) layout.Dimensions {
	return t.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
			// Title and refresh button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(t.theme, "Load Drawing")
						label.Font.Weight = 700
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(t.theme, &t.refreshListButton, "↻")
						btn.Background = color.NRGBA{R: 100, G: 100, B: 100, A: 200}
						return btn.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			// Saved files label
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body2(t.theme, "Saved files:")
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			// List of saved files
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(t.savedFiles) == 0 {
					label := material.Caption(t.theme, "No saved files found")
					label.Color = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
					return label.Layout(gtx)
				}

				// Show list of file buttons (max 5 visible with scroll)
				maxVisible := 5
				if len(t.savedFiles) > maxVisible {
					// TODO: Add scrolling support
				}

				return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
					func() []layout.FlexChild {
						var children []layout.FlexChild
						for i, filename := range t.savedFiles {
							// Capture loop variables
							idx := i
							name := filename
							children = append(children,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									btn := material.Button(t.theme, &t.fileButtons[idx], name)
									btn.Background = color.NRGBA{R: 70, G: 120, B: 200, A: 220}
									return btn.Layout(gtx)
								}),
							)
							if i < len(t.savedFiles)-1 {
								children = append(children, layout.Rigid(layout.Spacer{Height: 3}.Layout))
							}
						}
						return children
					}()...,
				)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			// Separator
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Caption(t.theme, "or enter filename:")
				label.Color = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			// Filename input
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Dp(200)
				gtx.Constraints.Max.X = gtx.Dp(200)
				editor := material.Editor(t.theme, &t.loadFilenameEditor, "")
				editor.TextSize = 14
				return editor.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			// Buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceStart}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(t.theme, &t.confirmLoadButton, "Load")
						btn.Background = color.NRGBA{R: 50, G: 100, B: 200, A: 255}
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 10}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(t.theme, &t.cancelLoadButton, "Cancel")
						btn.Background = color.NRGBA{R: 150, G: 50, B: 50, A: 255}
						return btn.Layout(gtx)
					}),
				)
			}),
		)
	})
}

// IsDialogOpen returns true if save or load dialog is open
func (t *Toolbar) IsDialogOpen() bool {
	return t.saveDialogOpen || t.loadDialogOpen
}

// refreshFileList updates the list of saved files
func (t *Toolbar) refreshFileList() {
	files, err := canvas.ListSavedFiles()
	if err != nil {
		// If error, clear the list
		t.savedFiles = nil
		t.fileButtons = nil
		return
	}

	t.savedFiles = files
	// Create/resize button slice to match file count
	if len(t.fileButtons) != len(files) {
		t.fileButtons = make([]widget.Clickable, len(files))
	}
}

// drawPanel is a helper that draws a white semi-transparent background with padding
func (t *Toolbar) drawPanel(gtx layout.Context, w layout.Widget) layout.Dimensions {
	return layout.Inset{Top: 10, Bottom: 10, Left: 10, Right: 10}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		backgroundColor := color.NRGBA{R: 255, G: 255, B: 255, A: 240}

		// Use Stack to layer background behind content
		return layout.Stack{}.Layout(gtx,
			// Background layer
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()
				paint.ColorOp{Color: backgroundColor}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				return layout.Dimensions{Size: gtx.Constraints.Min}
			}),
			// Content layer on top
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: 10, Bottom: 10, Left: 10, Right: 10}.Layout(gtx, w)
			}),
		)
	})
}

func (t *Toolbar) layoutColorSliders(gtx layout.Context) layout.Dimensions {
	// Constrain the width to 200dp
	gtx.Constraints.Max.X = gtx.Dp(200)

	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
		// Color preview
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			previewColor := color.NRGBA{
				R: uint8(t.redSlider.Value * 255),
				G: uint8(t.greenSlider.Value * 255),
				B: uint8(t.blueSlider.Value * 255),
				A: 255,
			}
			size := gtx.Dp(60)
			defer clip.Rect{Max: image.Pt(size, size)}.Push(gtx.Ops).Pop()
			paint.ColorOp{Color: previewColor}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			return layout.Dimensions{Size: image.Pt(size, size)}
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		// Red slider
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(t.theme, "R")
					label.Color = color.NRGBA{R: 255, G: 100, B: 100, A: 255}
					return label.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: 5}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Fixed width for slider
					gtx.Constraints.Min.X = gtx.Dp(150)
					gtx.Constraints.Max.X = gtx.Dp(150)
					slider := material.Slider(t.theme, &t.redSlider)
					slider.Color = color.NRGBA{R: 255, G: 100, B: 100, A: 255}
					return slider.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: 5}.Layout),
		// Green slider
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(t.theme, "G")
					label.Color = color.NRGBA{R: 100, G: 255, B: 100, A: 255}
					return label.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: 5}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Fixed width for slider
					gtx.Constraints.Min.X = gtx.Dp(150)
					gtx.Constraints.Max.X = gtx.Dp(150)
					slider := material.Slider(t.theme, &t.greenSlider)
					slider.Color = color.NRGBA{R: 100, G: 255, B: 100, A: 255}
					return slider.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: 5}.Layout),
		// Blue slider
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(t.theme, "B")
					label.Color = color.NRGBA{R: 100, G: 100, B: 255, A: 255}
					return label.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: 5}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Fixed width for slider
					gtx.Constraints.Min.X = gtx.Dp(150)
					gtx.Constraints.Max.X = gtx.Dp(150)
					slider := material.Slider(t.theme, &t.blueSlider)
					slider.Color = color.NRGBA{R: 100, G: 100, B: 255, A: 255}
					return slider.Layout(gtx)
				}),
			)
		}),
	)
}

func (t *Toolbar) layoutWidthSlider(gtx layout.Context) layout.Dimensions {
	// Constrain the width to 200dp
	gtx.Constraints.Max.X = gtx.Dp(200)

	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(t.theme, "Width")
			return label.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Fixed width for slider to make it bigger and more usable
			gtx.Constraints.Min.X = gtx.Dp(180)
			gtx.Constraints.Max.X = gtx.Dp(180)
			slider := material.Slider(t.theme, &t.widthSlider)
			return slider.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		// Width preview - show current width value
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			currentWidth := 2 + (t.widthSlider.Value * 18)
			label := material.Body2(t.theme, fmt.Sprintf("%.1f dp", currentWidth))
			return label.Layout(gtx)
		}),
	)
}
