# Surface Rendering Redesign

_Author: Codex_<br>
_Date: 2025-12-02_

This document describes a comprehensive redesign of Abacus’ rendering pipeline so every popup, toast, and base surface automatically receives the correct background without manual gap-filling. Instead of relying on string post-processing, we introduce explicit **Canvas** and **Surface** layers backed by Lip Gloss’ cell buffer to compose the final screen.

---

## Goals

1. **Guarantee background coverage** for every rendered cell (primary and secondary surfaces).
2. **Make overlays/toasts trivial** to implement: no filler spaces, no `fillBackground*` calls.
3. **Support true layering** (body + overlay + toast) using Z-order, without re-rendering the entire view as a single string.
4. **Leverage Lip Gloss v2 primitives** (styles, joins) but wrap them inside our own compositing buffer so resets are a non-issue.
5. **Incremental adoption**: we can migrate component-by-component while keeping the app usable.

Non-goals: replacing Bubble Tea, reworking theme definitions, or introducing animation.

---

## High-Level Architecture

```
┌───────────────────────────────────────────────────┐
│ AppView                                           │
│   ├── BaseLayer (primary canvas)                  │
│   │     ├── Header surface                        │
│   │     ├── Body surface (tree + details)         │
│   │     └── Footer surface                        │
│   ├── OverlayLayer (optional, secondary surface)  │
│   └── ToastLayer (optional, primary surface)      │
└───────────────────────────────────────────────────┘
```

Each “surface” renders into a **Canvas** object (width × height grid of cells with rune + style). Layers are composited front-to-back. Lip Gloss styles become convenience utilities for painting onto a canvas rather than directly mutating strings.

---

## Key Concepts

### 1. Canvas

Thin wrapper around [`github.com/charmbracelet/x/cellbuf`](https://pkg.go.dev/github.com/charmbracelet/x/cellbuf).

```go
type Canvas struct {
    buf cellbuf.Buffer
}

func NewCanvas(width, height int, bg lipgloss.TerminalColor) Canvas
func (c Canvas) Width() int
func (c Canvas) Height() int
func (c Canvas) Draw(x, y int, block string)
func (c Canvas) Render() string // returns ANSI frame
```

- `Draw` translates ANSI blocks (e.g., `lipgloss.NewStyle().Padding(...).Render("…")`) into cell writes.
- The constructor fills the entire canvas with the provided background once—no more per-line gap fixes.

### 2. Surface

Represents styling helpers plus a canonical `Canvas`.

```go
type Surface struct {
    Canvas Canvas
    Styles SurfaceStyles
}

type SurfaceStyles struct {
    Text          lipgloss.Style
    TextMuted     lipgloss.Style
    Accent        lipgloss.Style
    Error         lipgloss.Style
    Success       lipgloss.Style
    BorderNeutral lipgloss.Style
    BorderError   lipgloss.Style
}
```

- Surfaces produced via `NewPrimarySurface(width, height)` and `NewSecondarySurface(width, height)` (pull background colors from active theme).
- Styles already include the correct background, so components never touch global `styleFoo()` helpers.

### 3. Layers

`Layer` is just `type Layer interface { Render() Canvas }`. AppView gathers layers in order:

1. Base layer (always present)
2. Overlay layer (optional)
3. Toast layer(s) (optional; multiple to support stacked toasts later)

During `View()`, we composite:

```go
canvas := baseLayer.Render()
for _, layer := range overlays {
    canvas = canvas.Overlay(layer.Render()) // cell-wise overwrite where non-empty
}
return canvas.Render()
```

Overlay respects transparency: cells that are fully blank (space rune with zero width) do not overwrite the underlying canvas. This means we don’t need to worry about `Place()` erasing everything—the compositor only paints the actual popup region.

---

## Component API Changes

### Header / Body / Footer

- Replace `styleAppHeader().Render(...)` usage with drawing onto the primary surface directly.
- Build tree/detail panes by writing into sub-canvases and blitting them into the base canvas (keeps existing selection logic intact).

### Overlays

`StatusOverlay.View()` now returns a `Layer` implementation:

```go
func (m *StatusOverlay) Layer(width, height int) Layer {
    surf := NewSecondarySurface(width, height)
    box := surf.Styles.BorderNeutral.
        Border(lipgloss.RoundedBorder()).
        BorderForeground(theme.Current().BorderFocused()).
        Padding(1, 2).
        Width(overlayWidth)

    content := lipgloss.JoinVertical(lipgloss.Left,
        surf.Styles.Accent.Render(m.issueID),
        surf.Styles.TextMuted.Render(strings.Repeat("─", overlayWidth-4)),
        // …
    )

    popup := box.Render(content)

    return LayerFunc(func() Canvas {
        c := surf.Canvas.Clone()
        // center popup using coordinates
        x := (c.Width() - overlayWidth) / 2
        y := (c.Height() - overlayHeight) / 2
        c.Draw(x, y, popup)
        return c
    })
}
```

Overlay drawing no longer needs `fillSecondaryBackground` because the Canvas already has the background color baked in.

### Toasts

Same story, but they draw onto `NewPrimarySurface` and use `Canvas.Draw` with bottom-right alignment.

---

## Migration Strategy

1. **Infrastructure (Small)**
   - Introduce `canvas.go` (Canvas wrapper) and `surface.go`.
   - Provide `NewPrimarySurface` / `NewSecondarySurface`.
   - Add tests ensuring `Canvas.Draw` preserves background (per theme).

2. **Base Layer (Medium)**
   - Convert header/body/footer rendering to use a primary surface canvas.
   - Ensure we can re-create the current view purely with the new base canvas (no overlays yet).
   - Output should be byte-for-byte equivalent (or at least visually identical).

3. **Overlay Layer (Medium)**
   - Modify `View()` to accept `[]Layer` and composite them.
   - Port `StatusOverlay`, `LabelsOverlay`, `CreateOverlay`, `DeleteOverlay`, and `HelpOverlay` to the new `Layer` interface.
   - Remove `prepareOverlayContent`, `fillSecondaryBackground`, and all overlay-specific background hacks once the migration is complete.

4. **Toast Layer (Small)**
   - Port all toasts to return `Layer`s. Remove manual `overlayBottomRight` string logic once confirmed.

5. **Cleanup (Small)**
   - Delete unused global style helpers or rewrite them to delegate to surfaces.
   - Update `docs/UI_PRINCIPLES.md` with the new layering guidance.

---

## Testing Plan

1. **Unit Tests**
   - `canvas_test.go`: verify `Draw` respects ANSI, backgrounds, multi-line content.
   - `surface_test.go`: ensure style helpers include the correct background fore/back pairs.
2. **Golden Tests**
   - Capture ASCII snapshots of overlay/toast layers per theme (Dracula, Solarized, Nord) to guard against regressions.
3. **Integration Tests**
   - Extend `app_test.go` to assert that the final rendered string contains no `\x1b[0m ` sequences and no bare newlines without background codes.
4. **Performance Benchmarks**
   - Compare `View()` runtime before/after to ensure cell-buffer composition stays within budget (~2–3 ms per frame).

---

### Status (ab-smg0 – 2025-12-04)

- Added `TestOverlayAndToastGoldenSnapshots` plus six corresponding files under `testdata/ui/golden/` (Dracula/Solarized/Nord × overlay + toast). Refresh intentionally with `go test ./internal/ui -run TestOverlayAndToastGoldenSnapshots -update-golden`.
- Added `TestViewOmitsDefaultResetGaps` so the final `App.View()` output fails fast if `\x1b[0m ` ever sneaks back into the frame.
- Captured a baseline performance number via `BenchmarkAppViewLayering` (Apple M1 Max): **~1.76 ms/op (646 ops/s, 2.1 MB/op, 16.7K allocs)** which keeps us inside the 3 ms/frame budget. Notes recorded in `docs/SURFACE_RENDERING_SPIKE_NOTES.md`.
- Documented the Surface/Layer workflow and snapshot guardrails in `docs/UI_PRINCIPLES.md` so component authors know how to draw onto canvases without manual background hacks.

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| **Initial complexity**: Canvas/o layering logic is new. | Keep Canvas API tiny and well-tested. Document usage with examples. |
| **Lip Gloss updates**: relying on `x/cellbuf` ties us to Charmbracelet libs. | Already a transitive dependency via Bubble Tea; pin version. |
| **Migration churn**: tree/detail code touches many files. | Stage migration (base → overlays → toasts). Keep git history readable. |
| **Performance**: overlaying canvases might cost more than string concat. | Benchmark; cellbuf is written for this use case and should be plenty fast. |
| **Theme changes**: surfaces cache colors per render; switching theme mid-frame must rebuild all surfaces. | Surfaces created fresh each `View()` call. Avoid storing them on the struct. |

---

## Why Not “Better fillBackground”?

- Post-processing strings can only react to resets—it can’t detect gaps caused by `lipgloss.Place`, manual joins, or whitespace created after the fill.
- We would keep sprinkling `surf.Render(...)` wrappers everywhere; any missed call shows up as yet another visual glitch.
- The canvas approach moves background ownership to the lowest level (per cell), where it belongs. We never rely on resets because the buffer already knows each cell’s background color.

---

## Open Questions

1. **Do we need transparent overlays?** Right now, a blank cell in an overlay canvas “passes through” to the base layer. Do we ever need semi-transparent colors (probably not in ASCII world)? If yes, we could expand the cell metadata later.
2. **Should Surface expose custom styles (e.g., chips, badges)?** Probably yes; once the base migration is done, we can audit common patterns and add helpers to avoid re-implementing theme-aware styles.
3. **How to handle window resizing?** Canvas dimensions depend on the latest `tea.WindowSizeMsg`; we may want to reuse the base canvas between frames when only content changes, but that’s an optimization for later.

---

## Summary

By switching from ad-hoc string manipulation to explicit canvases and composited layers, we eliminate the entire class of “forgot to fill the background” bugs. Components stop worrying about ANSI resets and focus on content. The proposal keeps Lip Gloss front and center for styling but wraps it with a thin rendering layer tailored to Abacus’ needs. Once adopted, adding a new modal or toast should be as simple as drawing into the right surface—no magic incantations required.
