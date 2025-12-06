package ui

// Layer represents an overlay/toast that can render itself into a canvas
// matching the current terminal dimensions.
type Layer interface {
	Render() *Canvas
}

// LayerFunc is an adapter to allow ordinary functions to act as layers.
type LayerFunc func() *Canvas

// Render implements Layer for LayerFunc.
func (f LayerFunc) Render() *Canvas {
	return f()
}
