// Package chunking provides semantic and structural document chunking
// for the worker's ingestion pipeline.
package chunking

import "math"

// CosineSimilarity returns the cosine similarity between two vectors:
// dot(a, b) / (||a|| * ||b||). Range: [-1, 1]. Returns 0 if either vector
// is the zero vector or the lengths don't match.
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		da := float64(a[i])
		db := float64(b[i])
		dot += da * db
		na += da * da
		nb += db * db
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	return float32(dot / denom)
}

// CosineDistance returns 1 - CosineSimilarity. Range: [0, 2].
// Higher means more dissimilar. Returns 1 if either vector is zero.
func CosineDistance(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 1
	}
	sim := CosineSimilarity(a, b)
	return 1 - sim
}
