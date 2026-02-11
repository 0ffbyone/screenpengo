package input

import (
	"gioui.org/io/key"
	"gioui.org/layout"

	"screenpengo/internal/tool"
)

// Key name constants from Gio framework.
const (
	keyRed    = "R"
	keyGreen  = "G"
	keyBlue   = "B"
	keyYellow = "Y"
	keyOrange = "O"
	keyPink   = "P"
	keyBlur   = "X"
	keyThin   = "1"
	keyMedium = "2"
	keyThick  = "3"
	keyDim    = "A"
	keyClear  = "C"
)

// ActionType represents the type of action triggered by keyboard input.
type ActionType int

const (
	NoAction ActionType = iota
	SetColor
	SetWidth
	ToggleDim
	Clear
	Quit
)

// Action represents a user action triggered by keyboard input.
type Action struct {
	Type        ActionType
	ColorPreset tool.ColorPreset
	WidthPreset tool.WidthPreset
}

// KeyboardHandler processes keyboard events and returns actions.
type KeyboardHandler struct{}

// NewKeyboardHandler creates a new keyboard handler.
func NewKeyboardHandler() *KeyboardHandler {
	return &KeyboardHandler{}
}

// HandleEvents processes all pending keyboard events and returns actions to perform.
func (h *KeyboardHandler) HandleEvents(gtx layout.Context, keyTag *struct{}) []Action {
	var actions []Action

	// Consume focus events (required by Gio).
	for {
		_, ok := gtx.Event(key.FocusFilter{Target: keyTag})
		if !ok {
			break
		}
	}

	// Process key press events.
	for {
		ev, ok := gtx.Event(key.Filter{Focus: keyTag, Name: ""})
		if !ok {
			break
		}
		ke := ev.(key.Event)
		if ke.State != key.Press {
			continue
		}

		if action, ok := h.keyToAction(string(ke.Name)); ok {
			actions = append(actions, action)
		}
	}

	return actions
}

// keyToAction maps a key name to an action.
// Returns (action, true) if the key is recognized, or (zero action, false) otherwise.
func (h *KeyboardHandler) keyToAction(keyName string) (Action, bool) {
	switch keyName {
	case keyRed:
		return Action{Type: SetColor, ColorPreset: tool.Red}, true
	case keyGreen:
		return Action{Type: SetColor, ColorPreset: tool.Green}, true
	case keyBlue:
		return Action{Type: SetColor, ColorPreset: tool.Blue}, true
	case keyYellow:
		return Action{Type: SetColor, ColorPreset: tool.Yellow}, true
	case keyOrange:
		return Action{Type: SetColor, ColorPreset: tool.Orange}, true
	case keyPink:
		return Action{Type: SetColor, ColorPreset: tool.Pink}, true
	case keyBlur:
		return Action{Type: SetColor, ColorPreset: tool.Blur}, true
	case keyThin:
		return Action{Type: SetWidth, WidthPreset: tool.Thin}, true
	case keyMedium:
		return Action{Type: SetWidth, WidthPreset: tool.Medium}, true
	case keyThick:
		return Action{Type: SetWidth, WidthPreset: tool.Thick}, true
	case keyDim:
		return Action{Type: ToggleDim}, true
	case keyClear:
		return Action{Type: Clear}, true
	case string(key.NameEscape):
		return Action{Type: Quit}, true
	default:
		return Action{}, false
	}
}
