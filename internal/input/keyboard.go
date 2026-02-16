package input

import (
	"gioui.org/io/key"
	"gioui.org/layout"

	"screenpengo/internal/tool"
)

const (
	keyRed    = "R"
	keyGreen  = "G"
	keyBlue   = "B"
	keyYellow = "Y"
	keyOrange = "O"
	keyPink   = "P"
	keyBlur   = "X"
	keyEraser = "E"
	keyThin   = "1"
	keyMedium = "2"
	keyThick  = "3"
	keyDim    = "A"
	keyClear  = "C"
)

type ActionType int

const (
	NoAction ActionType = iota
	SetColor
	SetWidth
	ToggleDim
	Clear
	Quit
)

type Action struct {
	Type        ActionType
	ColorPreset tool.ColorPreset
	WidthPreset tool.WidthPreset
}

type KeyboardHandler struct{}

func NewKeyboardHandler() *KeyboardHandler {
	return &KeyboardHandler{}
}

func (h *KeyboardHandler) HandleEvents(gtx layout.Context, keyTag *struct{}) []Action {
	var actions []Action

	for {
		_, ok := gtx.Event(key.FocusFilter{Target: keyTag})
		if !ok {
			break
		}
	}

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
	case keyEraser:
		return Action{Type: SetColor, ColorPreset: tool.Eraser}, true
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
