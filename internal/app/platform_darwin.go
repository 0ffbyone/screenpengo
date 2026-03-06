package app

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework Cocoa
// #import <Cocoa/Cocoa.h>
//
// void setupOverlayWindow() {
//     dispatch_async(dispatch_get_main_queue(), ^{
//         @autoreleasepool {
//             for (NSWindow *window in [NSApplication sharedApplication].windows) {
//                 // Cover the full screen without entering fullscreen Space
//                 NSScreen *screen = [NSScreen mainScreen];
//                 if (screen) {
//                     [window setFrame:screen.frame display:YES animate:NO];
//                 }
//
//                 // Make window transparent
//                 [window setOpaque:NO];
//                 [window setBackgroundColor:[NSColor clearColor]];
//                 [window setHasShadow:NO];
//
//                 // Float above other windows, visible on all Spaces
//                 [window setLevel:NSFloatingWindowLevel];
//                 [window setCollectionBehavior:
//                     NSWindowCollectionBehaviorCanJoinAllSpaces |
//                     NSWindowCollectionBehaviorStationary];
//
//                 // Make the Metal layer transparent
//                 NSView *contentView = [window contentView];
//                 if (contentView) {
//                     contentView.wantsLayer = YES;
//                     if (contentView.layer) {
//                         contentView.layer.opaque = NO;
//                         contentView.layer.backgroundColor = NULL;
//                     }
//                 }
//
//                 [window makeKeyAndOrderFront:nil];
//             }
//         }
//     });
// }
import "C"
import "sync"

var platformOnce sync.Once

func initPlatform() {
	platformOnce.Do(func() {
		C.setupOverlayWindow()
	})
}
