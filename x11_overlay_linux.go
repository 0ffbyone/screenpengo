//go:build linux && !android

package main

/*
#cgo linux LDFLAGS: -lX11 -lXext
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/extensions/shape.h>
#include <stdlib.h>
#include <string.h>


static int last_xerr = 0;
static int focus_err_handler(Display* dpy, XErrorEvent* e) {
    (void)dpy;
    last_xerr = e->error_code;
    return 0;
}

static int safe_set_input_focus(Display* dpy, Window win) {
    int (*old)(Display*, XErrorEvent*) = XSetErrorHandler(focus_err_handler);
    last_xerr = 0;
    XSetInputFocus(dpy, win, RevertToPointerRoot, CurrentTime);
    XSync(dpy, False); // force error delivery now
    XSetErrorHandler(old);
    return last_xerr == 0;
}


static Atom atom(Display* dpy, const char* name) {
    return XInternAtom(dpy, name, False);
}

static int set_wm_state(Display* dpy, Window win, Atom stateAtom, int add) {
    Atom wmState = atom(dpy, "_NET_WM_STATE");
    if (wmState == None || stateAtom == None) return 0;

    XEvent e;
    memset(&e, 0, sizeof(e));
    e.xclient.type = ClientMessage;
    e.xclient.message_type = wmState;
    e.xclient.display = dpy;
    e.xclient.window = win;
    e.xclient.format = 32;
    e.xclient.data.l[0] = add ? 1 : 0; // _NET_WM_STATE_ADD / REMOVE
    e.xclient.data.l[1] = (long)stateAtom;
    e.xclient.data.l[2] = 0;
    e.xclient.data.l[3] = 1; // source indication: application
    e.xclient.data.l[4] = 0;

    Window root = DefaultRootWindow(dpy);
    long mask = SubstructureRedirectMask | SubstructureNotifyMask;
    return XSendEvent(dpy, root, False, mask, &e) != 0;
}

static int set_opacity(Display* dpy, Window win, unsigned long opacity) {
    Atom prop = atom(dpy, "_NET_WM_WINDOW_OPACITY");
    if (prop == None) return 0;
    return XChangeProperty(dpy, win, prop, XA_CARDINAL, 32, PropModeReplace,
                           (unsigned char*)&opacity, 1) == Success;
}


static int set_window_type_normal(Display* dpy, Window win) {
    Atom prop = atom(dpy, "_NET_WM_WINDOW_TYPE");
    Atom normal = atom(dpy, "_NET_WM_WINDOW_TYPE_NORMAL");
    if (prop == None || normal == None) return 0;
    return XChangeProperty(dpy, win, prop, XA_ATOM, 32, PropModeReplace,
                           (unsigned char*)&normal, 1) == Success;
}

static int set_click_through(Display* dpy, Window win, int enable) {
    int ev, er;
    if (!XShapeQueryExtension(dpy, &ev, &er)) return 0;
    if (enable) {
        XShapeCombineRectangles(dpy, win, ShapeInput, 0, 0, NULL, 0, ShapeSet, YXBanded);
    } else {
        XShapeCombineMask(dpy, win, ShapeInput, 0, 0, None, ShapeSet);
    }
    XFlush(dpy);
    return 1;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func x11EnableOverlayHints(display unsafe.Pointer, window uintptr) error {
	if display == nil || window == 0 {
		return fmt.Errorf("invalid X11 handles")
	}
	dpy := (*C.Display)(display)
	win := C.Window(window)

	above := C.atom(dpy, C.CString("_NET_WM_STATE_ABOVE"))
	full := C.atom(dpy, C.CString("_NET_WM_STATE_FULLSCREEN"))
	skipTaskbar := C.atom(dpy, C.CString("_NET_WM_STATE_SKIP_TASKBAR"))
	skipPager := C.atom(dpy, C.CString("_NET_WM_STATE_SKIP_PAGER"))

	C.set_window_type_normal(dpy, win)
	C.set_wm_state(dpy, win, above, 1)
	C.set_wm_state(dpy, win, full, 1)
	C.set_wm_state(dpy, win, skipTaskbar, 1)
	C.set_wm_state(dpy, win, skipPager, 1)

	C.safe_set_input_focus(dpy, win)
	C.XFlush(dpy)
	return nil
}

func x11SetOpacity(display unsafe.Pointer, window uintptr, opacity uint32) error {
	if display == nil || window == 0 {
		return fmt.Errorf("invalid X11 handles")
	}
	dpy := (*C.Display)(display)
	win := C.Window(window)
	if C.set_opacity(dpy, win, C.ulong(opacity)) == 0 {
		return fmt.Errorf("set opacity failed")
	}
	C.safe_set_input_focus(dpy, win)
	C.XFlush(dpy)
	return nil
}

func x11SetClickThrough(display unsafe.Pointer, window uintptr, enable bool) error {
	if display == nil || window == 0 {
		return fmt.Errorf("invalid X11 handles")
	}
	dpy := (*C.Display)(display)
	win := C.Window(window)
	var en C.int
	if enable {
		en = 1
	}
	if C.set_click_through(dpy, win, en) == 0 {
		return fmt.Errorf("shape extension not available")
	}
	return nil
}
