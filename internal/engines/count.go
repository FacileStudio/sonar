package engines

// countAt picks the smaller of two positive counts, falling back to a default.
// Shared by the keyed engines so their per-engine count config is honored.
func countAt(a, b int) int {
	switch {
	case b > 0:
		return b
	case a > 0:
		return a
	default:
		return 10
	}
}
