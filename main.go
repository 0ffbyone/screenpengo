package main

import (
	"image"
	"image/color"
	"log"
	"os"
	"runtime"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type Stroke struct {
	Col   color.NRGBA
	Width float32
	Pts   []f32.Point
}

type Annotator struct {
	keyTag struct{}
	ptrTag struct{}

	debug bool

	hasFocus bool

	curColor color.NRGBA
	width    float32
	dim      bool

	strokes []Stroke
	active  *Stroke
}

func main() {
	debug := os.Getenv("ANNOTATOR_DEBUG") != ""

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("starting gio-annotator (go=%s, os=%s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	log.Printf("env: XDG_SESSION_TYPE=%q WAYLAND_DISPLAY=%q DISPLAY=%q ANNOTATOR_DEBUG=%t",
		os.Getenv("XDG_SESSION_TYPE"), os.Getenv("WAYLAND_DISPLAY"), os.Getenv("DISPLAY"), debug)

	go func() {
		w := new(app.Window)
		w.Option(
			app.Size(900, 600),
			app.Title("gio-annotator"),
		)
		if err := run(w, debug); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	app.Main()
}

func run(w *app.Window, debug bool) error {
	a := &Annotator{
		debug:    debug,
		curColor: color.NRGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF},
		width:    6,
	}

	var ops op.Ops
	var frameN uint64

	for {
		evt := w.Event()
		switch e := evt.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			frameN++
			if frameN == 1 {
				log.Printf("first frame: size=%v, metric=%+v", e.Size, e.Metric)
				log.Printf("tip: click inside the window to focus it, then try keys: r/g/b, 1/2/3, c, a, esc")
			}

			keys, ptrs := a.handleInput(w, e)
			if a.debug && (keys > 0 || ptrs > 0) {
				log.Printf("frame=%d processed: key=%d pointer=%d focus=%v", frameN, keys, ptrs, a.hasFocus)
			}

			ops.Reset()
			a.layout(&ops, e)
			e.Frame(&ops)
		}
	}
}

func (a *Annotator) dlog(format string, args ...any) {
	if !a.debug {
		return
	}
	log.Printf(format, args...)
}

func (a *Annotator) handleInput(w *app.Window, e app.FrameEvent) (keys, pointers int) {

	// Keyboard events.
	for {
		ev, ok := e.Source.Event(
			key.FocusFilter{Target: &a.keyTag},
			key.Filter{Focus: &a.keyTag, Name: key.NameEscape},
			key.Filter{Focus: &a.keyTag, Name: "R"},
			key.Filter{Focus: &a.keyTag, Name: "G"},
			key.Filter{Focus: &a.keyTag, Name: "B"},
			key.Filter{Focus: &a.keyTag, Name: "K"},
			key.Filter{Focus: &a.keyTag, Name: "W"},
			key.Filter{Focus: &a.keyTag, Name: "A"},
			key.Filter{Focus: &a.keyTag, Name: "C"},
			key.Filter{Focus: &a.keyTag, Name: "1"},
			key.Filter{Focus: &a.keyTag, Name: "2"},
			key.Filter{Focus: &a.keyTag, Name: "3"},
			key.Filter{Focus: &a.keyTag, Name: ""},
		)
		if !ok {
			break
		}
		switch kev := ev.(type) {
		case key.FocusEvent:
			a.hasFocus = kev.Focus
			a.dlog("focus: %v", kev.Focus)
		case key.Event:
			// We only handle key presses.
			if kev.State != key.Press {
				continue
			}
			keys++
			a.dlog("key: name=%q mods=%v", kev.Name, kev.Modifiers)

			switch kev.Name {
			case key.NameEscape:
				w.Perform(system.ActionClose)
			case "R":
				a.curColor = color.NRGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}
				log.Printf("color=red")
			case "G":
				a.curColor = color.NRGBA{R: 0x00, G: 0xFF, B: 0x00, A: 0xFF}
				log.Printf("color=green")
			case "B":
				a.curColor = color.NRGBA{R: 0x00, G: 0x80, B: 0xFF, A: 0xFF}
				log.Printf("color=blue")
			case "K":
				a.curColor = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}
				log.Printf("color=black")
			case "W":
				a.curColor = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
				log.Printf("color=white")

			case "1":
				a.width = 2
				log.Printf("width=2")
			case "2":
				a.width = 6
				log.Printf("width=6")
			case "3":
				a.width = 12
				log.Printf("width=12")

			case "A":
				a.dim = !a.dim
				log.Printf("dim=%v", a.dim)
			case "C":
				a.strokes = nil
				a.active = nil
				log.Printf("clear")
			}
		}
	}

	// Pointer events.
	for {
		ev, ok := e.Source.Event(pointer.Filter{Target: &a.ptrTag, Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel})
		if !ok {
			break
		}
		pe, ok := ev.(pointer.Event)
		if !ok {
			continue
		}
		pointers++
		a.dlog("pointer: kind=%v pos=(%.1f,%.1f) buttons=%v", pe.Kind, pe.Position.X, pe.Position.Y, pe.Buttons)

		// Only primary button draws.
		if pe.Buttons != 0 && !pe.Buttons.Contain(pointer.ButtonPrimary) {
			continue
		}

		switch pe.Kind {
		case pointer.Press:
			// Click-to-focus: request focus when the user interacts.

			st := &Stroke{
				Col:   a.curColor,
				Width: a.width,
				Pts:   []f32.Point{pe.Position},
			}
			a.active = st
			log.Printf("stroke begin: col=%v width=%.1f", st.Col, st.Width)
		case pointer.Drag:
			if a.active != nil {
				pts := &a.active.Pts
				last := (*pts)[len(*pts)-1]
				dx := pe.Position.X - last.X
				dy := pe.Position.Y - last.Y
				// decimate points a bit
				if dx*dx+dy*dy >= 1.5*1.5 {
					*pts = append(*pts, pe.Position)
				}
			}
		case pointer.Release, pointer.Cancel:
			if a.active != nil {
				a.strokes = append(a.strokes, *a.active)
				a.active = nil
				log.Printf("stroke end: total=%d", len(a.strokes))
			}
		}
	}

	return keys, pointers
}

func (a *Annotator) layout(ops *op.Ops, e app.FrameEvent) {
	full := image.Rectangle{Max: e.Size}

	// Background.
	paint.FillShape(ops, color.NRGBA{R: 0xF5, G: 0xF5, B: 0xF5, A: 0xFF}, clip.Rect(full).Op())

	// Declare a full-window hit area and attach tags.
	area := clip.Rect(full).Push(ops)
	event.Op(ops, &a.ptrTag)
	event.Op(ops, &a.keyTag)
	// Ask Gio to route keyboard events to this tag.
	key.InputHintOp{Tag: &a.keyTag, Hint: key.HintAny}.Add(ops)
	// Request focus after the tag is in the input tree for this frame.
	if !a.hasFocus {
		e.Source.Execute(key.FocusCmd{Tag: &a.keyTag})
	}
	area.Pop()

	// Optional dim overlay.
	if a.dim {
		paint.FillShape(ops, color.NRGBA{A: 120}, clip.Rect(full).Op())
	}

	for i := range a.strokes {
		drawStroke(ops, a.strokes[i])
	}
	if a.active != nil {
		drawStroke(ops, *a.active)
		// Keep animating while drawing.
		e.Source.Execute(op.InvalidateCmd{})
	}
}

func drawStroke(ops *op.Ops, s Stroke) {
	if len(s.Pts) < 2 {
		return
	}
	var p clip.Path
	p.Begin(ops)
	p.MoveTo(s.Pts[0])
	for _, pt := range s.Pts[1:] {
		p.LineTo(pt)
	}
	path := p.End()
	shape := clip.Stroke{Path: path, Width: s.Width}.Op()
	paint.FillShape(ops, s.Col, shape)
}
