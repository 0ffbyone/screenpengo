//go:build linux && !android

package main

/*
#cgo linux LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <stdlib.h>

static int move_to_pointer(Display* dpy, Window win) {
    Window root = DefaultRootWindow(dpy);
    Window ret_root, ret_child;
    int root_x, root_y, win_x, win_y;
    unsigned int mask;
    if (!XQueryPointer(dpy, root, &ret_root, &ret_child, &root_x, &root_y, &win_x, &win_y, &mask)) {
        return 0;
    }
    // Move window so that its top-left is near the pointer position.
    // (Fullscreen will then pick this monitor on many WMs.)
    XMoveWindow(dpy, win, root_x - 50, root_y - 50);
    XFlush(dpy);
    return 1;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func x11MoveWindowToPointer(display unsafe.Pointer, window uintptr) error {
	if display == nil || window == 0 {
		return fmt.Errorf("invalid X11 handles")
	}
	dpy := (*C.Display)(display)
	win := C.Window(window)
	if C.move_to_pointer(dpy, win) == 0 {
		return fmt.Errorf("XQueryPointer failed")
	}
	return nil
}
