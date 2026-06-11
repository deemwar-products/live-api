package chunking

import (
	"math"
	"testing"
)

func TestCosineSimilarity_Identical(t *testing.T) {
	a := []float32{1, 2, 3}
	if got := CosineSimilarity(a, a); got != 1.0 {
		t.Errorf("identical: got %v want 1.0", got)
	}
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{0, 1}
	if got := CosineSimilarity(a, b); got != 0 {
		t.Errorf("orthogonal: got %v want 0", got)
	}
}

func TestCosineSimilarity_Opposite(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{-1, 0}
	if got := CosineSimilarity(a, b); got != -1 {
		t.Errorf("opposite: got %v want -1", got)
	}
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	a := []float32{0, 0, 0}
	b := []float32{1, 2, 3}
	if got := CosineSimilarity(a, b); got != 0 {
		t.Errorf("zero a: got %v want 0", got)
	}
	if got := CosineSimilarity(b, a); got != 0 {
		t.Errorf("zero b: got %v want 0", got)
	}
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	a := []float32{1, 2}
	b := []float32{1, 2, 3}
	if got := CosineSimilarity(a, b); got != 0 {
		t.Errorf("length mismatch: got %v want 0", got)
	}
}

func TestCosineSimilarity_Empty(t *testing.T) {
	if got := CosineSimilarity(nil, []float32{1}); got != 0 {
		t.Errorf("nil: got %v want 0", got)
	}
	if got := CosineSimilarity([]float32{}, []float32{}); got != 0 {
		t.Errorf("empty: got %v want 0", got)
	}
}

func TestCosineSimilarity_KnownValue(t *testing.T) {
	// a = (1, 2, 3), b = (4, 5, 6)
	// dot = 4+10+18 = 32
	// ||a|| = sqrt(14), ||b|| = sqrt(77)
	// sim = 32 / sqrt(14*77) = 32 / sqrt(1078)
	a := []float32{1, 2, 3}
	b := []float32{4, 5, 6}
	want := float32(32 / math.Sqrt(1078))
	got := CosineSimilarity(a, b)
	if math.Abs(float64(got-want)) > 1e-5 {
		t.Errorf("known: got %v want %v", got, want)
	}
}

func TestCosineDistance_Identical(t *testing.T) {
	a := []float32{1, 2, 3}
	if got := CosineDistance(a, a); got != 0 {
		t.Errorf("identical: got %v want 0", got)
	}
}

func TestCosineDistance_Orthogonal(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{0, 1}
	if got := CosineDistance(a, b); got != 1 {
		t.Errorf("orthogonal: got %v want 1", got)
	}
}

func TestCosineDistance_Opposite(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{-1, 0}
	if got := CosineDistance(a, b); got != 2 {
		t.Errorf("opposite: got %v want 2", got)
	}
}

func TestCosineDistance_ZeroVector(t *testing.T) {
	a := []float32{0, 0}
	b := []float32{1, 1}
	if got := CosineDistance(a, b); got != 1 {
		t.Errorf("zero: got %v want 1", got)
	}
}

func TestCosineDistance_LengthMismatch(t *testing.T) {
	a := []float32{1}
	b := []float32{1, 2}
	if got := CosineDistance(a, b); got != 1 {
		t.Errorf("mismatch: got %v want 1", got)
	}
}
