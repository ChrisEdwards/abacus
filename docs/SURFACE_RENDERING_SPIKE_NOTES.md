# Surface Rendering Spike Notes (ab-wjzw)

Date: 2025-12-03  
Author: ChartreuseCastle

## What was prototyped

- Added a thin `Canvas` helper around `github.com/charmbracelet/x/cellbuf` that allocates an off-screen buffer per `View()` call.
- Base header/body/footer strings are still produced with Lip Gloss, but are now written into the buffer with `ScreenWriter.PrintCropAt`.
- Modal overlays use a new buffer overlay helper that centers content with configurable margins (top=header, bottom=footer) and a bottom-right helper for overlay error toasts.
- The final frame is emitted by calling `cellbuf.Render` on the backing screen after converting CRLF → LF for Bubble Tea.

## Key findings

1. **PrintCropAt requires CRLF normalization**: To keep each ANSI line aligned, newline sequences need to be rewritten to `\r\n` before writing into the buffer. This is hidden inside `Canvas.DrawStringAt`.
2. **ScreenWriter needs a real Screen**: There isn’t a buffer-specific writer. Creating a throwaway `cellbuf.Screen` backed by `io.Discard` works well for composing frames before returning them as strings.
3. **Rendering to Bubble Tea needs LF**: `cellbuf.Render` produces CRLF sequences; Bubble Tea expects LF-only strings. Converting before returning avoids double-spacing artifacts.

## Open questions / TODOs

1. Introduce shared `Surface` style sets so components stop calling the legacy `styleFoo` helpers directly.
2. Port toast overlays to the Canvas helpers so we can delete the string-based `overlayBottomRight` path.
3. Create unit/golden tests that snapshot the centered overlay output per theme once styles move fully onto surfaces.
