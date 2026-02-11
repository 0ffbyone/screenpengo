package input

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
)

const (
	noButtons pointer.Buttons = 0
)

// PointerActionType represents the type of action triggered by pointer input.
type PointerActionType int

const (
	NoPointerAction PointerActionType = iota
	StartStroke
	AddPoint
	FinishStroke
	MoveCursor
)

// PointerAction represents a user action triggered by pointer input.
type PointerAction struct {
	Type     PointerActionType
	Position f32.Point
}

// PointerHandler processes pointer events and returns actions.
type PointerHandler struct{}

// NewPointerHandler creates a new pointer handler.
func NewPointerHandler() *PointerHandler {
	return &PointerHandler{}
}

// HandleEvents processes all pending pointer events and returns actions to perform.
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
			// Only handle primary button (left mouse button).
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
			// Track cursor position for visual feedback
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
