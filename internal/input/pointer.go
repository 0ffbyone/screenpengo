package input

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
)

const (
	noButtons pointer.Buttons = 0
)

type PointerActionType int

const (
	NoPointerAction PointerActionType = iota
	StartStroke
	AddPoint
	FinishStroke
	MoveCursor
)

type PointerAction struct {
	Type     PointerActionType
	Position f32.Point
}

type PointerHandler struct{}

func NewPointerHandler() *PointerHandler {
	return &PointerHandler{}
}

func (h *PointerHandler) HandleEvents(gtx layout.Context, ptrTag *struct{}) []PointerAction {
	var actions []PointerAction

	for {
		ev, ok := gtx.Event(pointer.Filter{
			Target: ptrTag,
			Kinds:  pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel | pointer.Move,
		})
		if !ok {
			break
		}
		pe := ev.(pointer.Event)

		switch pe.Kind {
		case pointer.Press:
			isPrimaryButton := pe.Buttons&pointer.ButtonPrimary != noButtons
			if isPrimaryButton {
				actions = append(actions, PointerAction{
					Type:     StartStroke,
					Position: pe.Position,
				})
			}
		case pointer.Drag:
			actions = append(actions, PointerAction{
				Type:     AddPoint,
				Position: pe.Position,
			})
		case pointer.Move:
			actions = append(actions, PointerAction{
				Type:     MoveCursor,
				Position: pe.Position,
			})
		case pointer.Release, pointer.Cancel:
			actions = append(actions, PointerAction{
				Type: FinishStroke,
			})
		}
	}

	return actions
}
