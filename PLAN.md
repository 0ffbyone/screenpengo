# Refactoring Plan: screenpengo to Idiomatic Go

## Context

The screenpengo project is currently a 250-line monolithic `main.go` file where the `Annotator` struct handles all responsibilities: drawing state, event handling, rendering, and configuration. This violates separation of concerns and makes the code hard to test and maintain.

**Current problems:**
- God object: Annotator does 12+ different things
- No testability: All logic coupled to Gio UI framework
- Hard-coded configuration: Colors and widths embedded in switch statements
- No package organization: Everything in `package main`
- Module name mismatch: `annotator_gio_fullscreen_clean` vs actual usage

**Goal:** Refactor to idiomatic Go with clear separation of concerns while maintaining all functionality.

## Target Package Structure

```
screenpengo/
├── cmd/
│   └── screenpen-go/
│       └── main.go                    # Entry point (Gio window setup)
├── internal/
│   ├── app/
│   │   └── app.go                     # Application coordinator
│   ├── canvas/
│   │   ├── canvas.go                  # Canvas type, stroke management
│   │   └── stroke.go                  # Stroke type, drawing data
│   ├── tool/
│   │   └── pen.go                     # PenConfig type
│   ├── input/
│   │   ├── keyboard.go                # Keyboard handler
│   │   └── pointer.go                 # Pointer handler
│   └── render/
│       └── gio.go                     # Gio-specific rendering
├── go.mod                             # Rename module to "screenpengo"
└── build-linux.sh                     # Update binary name
```

## Extracted Types and Responsibilities

### 1. Canvas Package (`internal/canvas/`)

**stroke.go:**
```go
type Stroke struct {
    Points []f32.Point
    Color  color.NRGBA
    Width  float32
}
```

**canvas.go:**
```go
type Canvas struct {
    strokes []Stroke
    current *Stroke
}

// Methods:
// - StartStroke(color, width) - begins new stroke
// - AddPoint(pt) - adds point to current stroke
// - FinishStroke() - commits current stroke
// - Clear() - clears all strokes
// - Strokes() []Stroke - returns all strokes
// - CurrentStroke() *Stroke - returns current stroke
```

**Responsibility:** Manages drawing data. Pure domain logic, no Gio dependency.

### 2. Tool Package (`internal/tool/`)

**pen.go:**
```go
type PenConfig struct {
    Color   color.NRGBA
    WidthDp float32
}

// Methods:
// - SetColor(c color.NRGBA)
// - SetWidth(dp float32)
// - ColorPreset(name string) color.NRGBA - returns predefined colors
// - WidthPreset(level int) float32 - returns 1/2/3 widths
```

**Responsibility:** Tool configuration. Knows about color palette and width presets.

### 3. Input Package (`internal/input/`)

**keyboard.go:**
```go
type KeyboardHandler struct {
    pen *tool.PenConfig
}

// HandleKey(name string, modifiers key.Modifiers) Action
// Returns: ColorChange, WidthChange, ToggleDim, Clear, Quit actions
```

**pointer.go:**
```go
type PointerHandler struct{}

// ProcessEvent(kind pointer.Kind, pos f32.Point, buttons pointer.Buttons) PointerAction
// Returns: StartStroke, AddPoint, EndStroke actions
```

**Responsibility:** Translates input events to actions. No direct state mutation.

### 4. Render Package (`internal/render/`)

**gio.go:**
```go
type GioRenderer struct {
    dim bool
}

// RenderFrame(gtx layout.Context, canvas *canvas.Canvas)
// renderStroke(ops *op.Ops, s *canvas.Stroke) - internal helper
```

**Responsibility:** Gio-specific rendering. Draws strokes and background.

### 5. App Package (`internal/app/`)

**app.go:**
```go
type App struct {
    canvas   *canvas.Canvas
    pen      *tool.PenConfig
    keyboard *input.KeyboardHandler
    pointer  *input.PointerHandler
    renderer *render.GioRenderer

    keyTag struct{}
    ptrTag struct{}
    debug  bool
}

// Frame(gtx layout.Context) - main frame handler
// Coordinates between input handlers, canvas, and renderer
```

**Responsibility:** Application coordinator. Owns all components, handles Gio event loop.

## Refactoring Steps (Incremental with Testing)

**IMPORTANT:** After each step, we will build and test the application to ensure nothing breaks before proceeding to the next step.

---

### Step 1: Prepare Structure ✓ TEST CHECKPOINT
1. Rename module in `go.mod` from `annotator_gio_fullscreen_clean` to `screenpengo`
2. Update imports in `main.go` to use `screenpengo`
3. Run `go mod tidy` to verify

**Test:** `go build && ./screenpen-go` - should work exactly as before

---

### Step 2: Extract Canvas Package ✓ TEST CHECKPOINT
1. Create `internal/canvas/stroke.go`:
   - Move Stroke type from main.go
   - Add `appendInterpolated` helper function
2. Create `internal/canvas/canvas.go`:
   - Create Canvas type with `strokes []Stroke` and `current *Stroke`
   - Add methods: StartStroke, AddPoint, FinishStroke, Clear, Strokes, CurrentStroke
3. Update `main.go`:
   - Import `screenpengo/internal/canvas`
   - Replace Annotator's stroke fields with `canvas *canvas.Canvas`
   - Update all stroke operations to use canvas methods

**Why first:** Canvas is pure domain logic, easiest to extract, highest value for testability.

**Test:** `go build && ./screenpen-go`
- Verify drawing works with mouse
- Verify C key clears drawings
- Verify Esc quits

---

### Step 3: Extract Pen Configuration ✓ TEST CHECKPOINT
1. Create `internal/tool/pen.go`:
   - Create PenConfig type with Color and WidthDp fields
   - Add color constants (Red, Green, Blue, Yellow, Orange, Pink, Blur)
   - Add ColorPreset(name string) method
   - Add WidthPreset(level int) method
2. Update `main.go`:
   - Import `screenpengo/internal/tool`
   - Replace Annotator's col/widthDp fields with `pen *tool.PenConfig`
   - Update keyboard handler to use pen.ColorPreset() and pen.WidthPreset()

**Why second:** Simple value object with no dependencies.

**Test:** `go build && ./screenpen-go`
- Verify R/G/B/Y/O/P color changes work
- Verify X blur pen works
- Verify 1/2/3 width changes work

---

### Step 4: Extract Keyboard Handler ✓ TEST CHECKPOINT
1. Create `internal/input/keyboard.go`:
   - Create KeyboardHandler type
   - Move handleKeys logic, return action types (ColorChange, WidthChange, ToggleDim, Clear, Quit)
2. Update `main.go`:
   - Import `screenpengo/internal/input`
   - Replace handleKeys() with keyboard handler
   - Apply actions returned by handler

**Why third:** Separates input translation from state mutation.

**Test:** `go build && ./screenpen-go`
- Verify all keyboard shortcuts still work (R/G/B/Y/O/P/X/1/2/3/A/C/Esc)
- Verify A key toggles dimming
- Verify debug logging shows key events if ANNOTATOR_DEBUG=1

---

### Step 5: Extract Pointer Handler ✓ TEST CHECKPOINT
1. Create `internal/input/pointer.go`:
   - Create PointerHandler type
   - Move handlePointer logic, return actions (StartStroke, AddPoint, EndStroke)
2. Update `main.go`:
   - Replace handlePointer() with pointer handler
   - Apply actions returned by handler

**Why fourth:** Completes input separation.

**Test:** `go build && ./screenpen-go`
- Verify drawing with press/drag/release works
- Verify smooth interpolated lines
- Verify debug logging shows pointer events if ANNOTATOR_DEBUG=1

---

### Step 6: Extract Renderer ✓ TEST CHECKPOINT
1. Create `internal/render/gio.go`:
   - Create GioRenderer type with dim field
   - Move drawStroke → renderStroke method
   - Move dpToPx helper
   - Add RenderFrame method (background, dimming, strokes)
2. Update `main.go`:
   - Import `screenpengo/internal/render`
   - Replace direct rendering in frame() with renderer.RenderFrame()

**Why fifth:** Isolates Gio rendering concerns.

**Test:** `go build && ./screenpen-go`
- Verify all drawing appears correctly
- Verify dimming (A key) renders properly
- Verify transparency works

---

### Step 7: Create App Coordinator ✓ TEST CHECKPOINT
1. Create `internal/app/app.go`:
   - Create App type that owns canvas, pen, keyboard, pointer, renderer
   - Move frame() method → App.Frame()
   - App coordinates between all components
2. Update `main.go`:
   - Create App instance
   - Delegate to App.Frame() in event loop

**Why sixth:** Central coordinator with all dependencies.

**Test:** `go build && ./screenpen-go`
- Full smoke test of all features

---

### Step 8: Clean Entry Point ✓ TEST CHECKPOINT
1. Create `cmd/screenpen-go/main.go`:
   - Minimal main function
   - Create App, start Gio window
2. Update `build-linux.sh` to build from `./cmd/screenpen-go`
3. Remove old `main.go` from root

**Why last:** Entry point is trivial once everything else works.

**Final Test:** `./build-linux.sh && ./screenpen-go`
- Complete feature verification

---

**After EACH step:** User will test the application to confirm it still works before we proceed to the next step.

## Critical Files to Modify

- **Source:** `/Users/anton.ermakov/dev/my/git-clones/screenpengo/main.go`
- **Module:** `/Users/anton.ermakov/dev/my/git-clones/screenpengo/go.mod`
- **Build script:** `/Users/anton.ermakov/dev/my/git-clones/screenpengo/build-linux.sh`

## Verification Strategy

After each step:
1. Run `go build ./...` to verify compilation
2. Run the app and test basic drawing
3. Test keyboard shortcuts (R/G/B colors, 1/2/3 widths, C clear, Esc quit)
4. Test dimming (A key)

Final verification:
1. Build: `./build-linux.sh`
2. Run: `ANNOTATOR_DEBUG=1 ./screenpen-go`
3. Verify all features work:
   - Drawing with mouse/pointer
   - Color changes (R, G, B, Y, O, P, X)
   - Width changes (1, 2, 3)
   - Dimming toggle (A)
   - Clear (C)
   - Quit (Esc)
4. Check debug logging shows pointer and key events

## Future Testability

With this structure, these tests become possible:
- Unit tests for Canvas (stroke management without Gio)
- Unit tests for PenConfig (color/width presets)
- Unit tests for input handlers (event → action mapping)
- Integration tests for App (action → state changes)

## Trade-offs and Decisions

**What we're NOT doing:**
- Not adding tests in this refactoring (that's a separate effort)
- Not bringing back X11 platform code yet (can be added later with proper interfaces)
- Not adding configuration files (keep hard-coded presets for now)
- Not over-engineering with complex abstractions

**What we ARE doing:**
- Clear separation of concerns
- Testable domain logic
- Idiomatic Go package structure
- Maintainable code for future enhancements

## Build Script Changes

Update `build-linux.sh` to build new entry point:
```bash
#!/bin/bash
set -e
go mod tidy
go build -o screenpen-go ./cmd/screenpen-go
```
