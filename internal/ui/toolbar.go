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

type Toolbar struct {
	colorButton  widget.Clickable
	widthButton  widget.Clickable
	eraserButton widget.Clickable
	shapesButton widget.Clickable
	saveButton   widget.Clickable
	loadButton   widget.Clickable

	circleButton    widget.Clickable
	rectangleButton widget.Clickable
	lineButton      widget.Clickable
	arrowButton     widget.Clickable

	confirmSaveButton widget.Clickable
	cancelSaveButton  widget.Clickable

	confirmLoadButton widget.Clickable
	cancelLoadButton  widget.Clickable

	filenameEditor     widget.Editor
	loadFilenameEditor widget.Editor

	savedFiles        []string
	fileButtons       []widget.Clickable
	refreshListButton widget.Clickable

	redSlider   widget.Float
	greenSlider widget.Float
	blueSlider  widget.Float

	widthSlider widget.Float

	prevRedValue   float32
	prevGreenValue float32
	prevBlueValue  float32
	prevWidthValue float32

	colorPickerOpen  bool
	widthPickerOpen  bool
	shapesPickerOpen bool
	saveDialogOpen   bool
	loadDialogOpen   bool

	eraserActive bool

	theme *material.Theme
}

func NewToolbar(theme *material.Theme) *Toolbar {
	saveEditor := widget.Editor{
		SingleLine: true,
		Submit:     true,
	}
	saveEditor.SetText("")

	loadEditor := widget.Editor{
		SingleLine: true,
		Submit:     true,
	}
	loadEditor.SetText("")

	return &Toolbar{
		theme:              theme,
		redSlider:          widget.Float{Value: 1.0},
		greenSlider:        widget.Float{Value: 0.0},
		blueSlider:         widget.Float{Value: 0.0},
		widthSlider:        widget.Float{Value: 0.5},
		filenameEditor:     saveEditor,
		loadFilenameEditor: loadEditor,
	}
}

func (t *Toolbar) HandleEvents(gtx layout.Context) (currentColor color.NRGBA, currentWidth float32, eraserClicked bool, slidersChanged bool, selectedShape tool.ShapeType, saveRequested bool, saveFilename string, loadRequested bool, loadFilename string) {
	selectedShape = tool.NoShape
	saveRequested = false
	loadRequested = false
	if t.colorButton.Clicked(gtx) {
		t.colorPickerOpen = !t.colorPickerOpen
		if t.colorPickerOpen {
			t.widthPickerOpen = false
		}
	}

	if t.widthButton.Clicked(gtx) {
		t.widthPickerOpen = !t.widthPickerOpen
		if t.widthPickerOpen {
			t.colorPickerOpen = false
			t.shapesPickerOpen = false
		}
	}

	if t.shapesButton.Clicked(gtx) {
		t.shapesPickerOpen = !t.shapesPickerOpen
		if t.shapesPickerOpen {
			t.colorPickerOpen = false
			t.widthPickerOpen = false
		}
	}

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

	if t.saveButton.Clicked(gtx) {
		t.saveDialogOpen = !t.saveDialogOpen
		if t.saveDialogOpen {
			t.colorPickerOpen = false
			t.widthPickerOpen = false
			t.shapesPickerOpen = false
			t.loadDialogOpen = false
		}
	}

	if t.confirmSaveButton.Clicked(gtx) {
		saveRequested = true
		saveFilename = t.filenameEditor.Text()
		t.saveDialogOpen = false
	}
	if t.cancelSaveButton.Clicked(gtx) {
		t.saveDialogOpen = false
	}

	if t.loadButton.Clicked(gtx) {
		t.loadDialogOpen = !t.loadDialogOpen
		if t.loadDialogOpen {
			t.colorPickerOpen = false
			t.widthPickerOpen = false
			t.shapesPickerOpen = false
			t.saveDialogOpen = false
			t.refreshFileList()
		}
	}

	if t.refreshListButton.Clicked(gtx) {
		t.refreshFileList()
	}

	for i := range t.fileButtons {
		if t.fileButtons[i].Clicked(gtx) {
			loadRequested = true
			loadFilename = t.savedFiles[i]
			t.loadDialogOpen = false
			break
		}
	}

	if t.confirmLoadButton.Clicked(gtx) {
		loadRequested = true
		loadFilename = t.loadFilenameEditor.Text()
		t.loadDialogOpen = false
	}
	if t.cancelLoadButton.Clicked(gtx) {
		t.loadDialogOpen = false
	}

	if t.eraserButton.Clicked(gtx) {
		t.eraserActive = !t.eraserActive

		if t.eraserActive {
			eraserClicked = true
			t.colorPickerOpen = false
			t.widthPickerOpen = false
		} else {
			slidersChanged = true
		}
	}

	if t.redSlider.Value != t.prevRedValue ||
		t.greenSlider.Value != t.prevGreenValue ||
		t.blueSlider.Value != t.prevBlueValue ||
		t.widthSlider.Value != t.prevWidthValue {
		slidersChanged = true
		t.eraserActive = false
		t.prevRedValue = t.redSlider.Value
		t.prevGreenValue = t.greenSlider.Value
		t.prevBlueValue = t.blueSlider.Value
		t.prevWidthValue = t.widthSlider.Value
	}

	currentColor = color.NRGBA{
		R: uint8(t.redSlider.Value * 255),
		G: uint8(t.greenSlider.Value * 255),
		B: uint8(t.blueSlider.Value * 255),
		A: 255,
	}

	currentWidth = 2 + (t.widthSlider.Value * 18)

	return currentColor, currentWidth, eraserClicked, slidersChanged, selectedShape, saveRequested, saveFilename, loadRequested, loadFilename
}

func (t *Toolbar) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, layout.Spacer{}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return t.layoutMainButtons(gtx)
				}),
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

func (t *Toolbar) layoutMainButtons(gtx layout.Context) layout.Dimensions {
	return t.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(t.theme, &t.colorButton, "Color")
				btn.Background = color.NRGBA{R: 70, G: 70, B: 70, A: 220}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(t.theme, &t.widthButton, "Width")
				btn.Background = color.NRGBA{R: 70, G: 70, B: 70, A: 220}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(t.theme, &t.eraserButton, "Eraser")
				if t.eraserActive {
					btn.Background = color.NRGBA{R: 100, G: 180, B: 255, A: 255}
				} else {
					btn.Background = color.NRGBA{R: 220, G: 220, B: 220, A: 220}
				}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(t.theme, &t.shapesButton, "Shapes")
				btn.Background = color.NRGBA{R: 70, G: 70, B: 70, A: 220}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(t.theme, &t.saveButton, "Save")
				btn.Background = color.NRGBA{R: 50, G: 150, B: 50, A: 220}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(t.theme, &t.loadButton, "Load")
				btn.Background = color.NRGBA{R: 50, G: 100, B: 200, A: 220}
				return btn.Layout(gtx)
			}),
		)
	})
}

func (t *Toolbar) layoutColorPicker(gtx layout.Context) layout.Dimensions {
	return t.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return t.layoutColorSliders(gtx)
	})
}

func (t *Toolbar) layoutWidthPicker(gtx layout.Context) layout.Dimensions {
	return t.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return t.layoutWidthSlider(gtx)
	})
}

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

func (t *Toolbar) layoutSaveDialog(gtx layout.Context) layout.Dimensions {
	return t.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(t.theme, "Save Drawing")
				label.Font.Weight = 700
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body2(t.theme, "Filename:")
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Dp(200)
				gtx.Constraints.Max.X = gtx.Dp(200)

				editor := material.Editor(t.theme, &t.filenameEditor, "Enter filename...")
				editor.TextSize = 14
				return editor.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Caption(t.theme, "Saved to ~/.screenpen/")
				label.Color = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 15}.Layout),
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

func (t *Toolbar) layoutLoadDialog(gtx layout.Context) layout.Dimensions {
	return t.drawPanel(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
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
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Body2(t.theme, "Saved files:")
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(t.savedFiles) == 0 {
					label := material.Caption(t.theme, "No saved files found")
					label.Color = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
					return label.Layout(gtx)
				}

				return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
					func() []layout.FlexChild {
						var children []layout.FlexChild
						for i, filename := range t.savedFiles {
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
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Caption(t.theme, "or enter filename:")
				label.Color = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Dp(200)
				gtx.Constraints.Max.X = gtx.Dp(200)
				editor := material.Editor(t.theme, &t.loadFilenameEditor, "")
				editor.TextSize = 14
				return editor.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
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

func (t *Toolbar) IsDialogOpen() bool {
	return t.saveDialogOpen || t.loadDialogOpen
}

func (t *Toolbar) refreshFileList() {
	files, err := canvas.ListSavedFiles()
	if err != nil {
		t.savedFiles = nil
		t.fileButtons = nil
		return
	}

	t.savedFiles = files
	if len(t.fileButtons) != len(files) {
		t.fileButtons = make([]widget.Clickable, len(files))
	}
}

func (t *Toolbar) drawPanel(gtx layout.Context, w layout.Widget) layout.Dimensions {
	return layout.Inset{Top: 10, Bottom: 10, Left: 10, Right: 10}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		backgroundColor := color.NRGBA{R: 255, G: 255, B: 255, A: 240}

		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()
				paint.ColorOp{Color: backgroundColor}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				return layout.Dimensions{Size: gtx.Constraints.Min}
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: 10, Bottom: 10, Left: 10, Right: 10}.Layout(gtx, w)
			}),
		)
	})
}

func (t *Toolbar) layoutColorSliders(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Max.X = gtx.Dp(200)

	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
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
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(t.theme, "R")
					label.Color = color.NRGBA{R: 255, G: 100, B: 100, A: 255}
					return label.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: 5}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = gtx.Dp(150)
					gtx.Constraints.Max.X = gtx.Dp(150)
					slider := material.Slider(t.theme, &t.redSlider)
					slider.Color = color.NRGBA{R: 255, G: 100, B: 100, A: 255}
					return slider.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: 5}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(t.theme, "G")
					label.Color = color.NRGBA{R: 100, G: 255, B: 100, A: 255}
					return label.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: 5}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = gtx.Dp(150)
					gtx.Constraints.Max.X = gtx.Dp(150)
					slider := material.Slider(t.theme, &t.greenSlider)
					slider.Color = color.NRGBA{R: 100, G: 255, B: 100, A: 255}
					return slider.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: 5}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(t.theme, "B")
					label.Color = color.NRGBA{R: 100, G: 100, B: 255, A: 255}
					return label.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: 5}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
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
	gtx.Constraints.Max.X = gtx.Dp(200)

	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(t.theme, "Width")
			return label.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Dp(180)
			gtx.Constraints.Max.X = gtx.Dp(180)
			slider := material.Slider(t.theme, &t.widthSlider)
			return slider.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			currentWidth := 2 + (t.widthSlider.Value * 18)
			label := material.Body2(t.theme, fmt.Sprintf("%.1f dp", currentWidth))
			return label.Layout(gtx)
		}),
	)
}
