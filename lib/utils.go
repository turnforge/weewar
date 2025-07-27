package weewar

// max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Simple math helpers (since we can't import math in WASM easily)
func cos(angle float64) float64 {
	// Simple cosine approximation using Taylor series
	// cos(x) ≈ 1 - x²/2! + x⁴/4! - x⁶/6!
	x := angle
	for x > 3.14159*2 {
		x -= 3.14159 * 2
	}
	for x < 0 {
		x += 3.14159 * 2
	}

	x2 := x * x
	return 1 - x2/2 + x2*x2/24 - x2*x2*x2/720
}

func sin(angle float64) float64 {
	// sin(x) = cos(x - π/2)
	return cos(angle - 3.14159/2)
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper math functions
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func approximateCos(angle float64) float64 {
	// Simple approximation - in a real implementation, use math.Cos
	return 1.0 - angle*angle/2.0 + angle*angle*angle*angle/24.0
}

func approximateSin(angle float64) float64 {
	// Simple approximation - in a real implementation, use math.Sin
	return angle - angle*angle*angle/6.0 + angle*angle*angle*angle*angle/120.0
}
