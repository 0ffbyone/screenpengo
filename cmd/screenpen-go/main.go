package main

import (
	"sync"

	"gioui.org/app"
	"gioui.org/op"

	internalApp "screenpengo/internal/app"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		w := new(app.Window)
		w.Option(
			app.Title("gio-screenpen"),
			app.Decorated(false),
			app.Fullscreen.Option(),
		)

		a := internalApp.New()

		var ops op.Ops
		for {
			switch e := w.Event().(type) {
			case app.DestroyEvent:
				return
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				a.Frame(gtx)
				e.Frame(gtx.Ops)
			}
		}
	}()

	app.Main()
	wg.Wait()
}
