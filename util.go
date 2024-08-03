package main

import (
	"math"
	"math/rand"
)

// Generates a random integer between min and max
func randIntInRange(min int, max int, r *rand.Rand) int {
	return r.Intn(max-min) + min
}

// Generates a random float between min and max
func randFloatInRange(min float64, max float64, r *rand.Rand) float64 {
	return min + r.Float64()*(max-min)
}

// Calculates the straight line distance between two points
func pythagDistance(x1 int, y1 int, x2 int, y2 int) int {
	distanceX := math.Abs(float64(x2 - x1))
	distanceY := math.Abs(float64(y2 - y1))
	return int(math.Sqrt(math.Pow(distanceX, 2) + math.Pow(distanceY, 2)))
}
