package formula

import (
	"math"
	"testing"
)

func approx(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

// ============================================================================
// Core formulas
// ============================================================================

func TestCalculateRisk_Equation1(t *testing.T) {
	// PDF p.11 Eq.1: Risk = (H × V × LCC)^(1/3)
	// Worked example: H=5.2, V=4.8, LCC=6.1 → Risk ≈ 5.3 (PDF example, rounded to 1dp)
	r := CalculateRisk(5.2, 4.8, 6.1)
	if !approx(r, 5.3, 0.05) {
		t.Errorf("Risk(5.2,4.8,6.1) = %v, want ≈5.3", r)
	}
}

func TestCalculateRisk_ZeroDimension(t *testing.T) {
	if r := CalculateRisk(0, 5, 5); r != 0 {
		t.Errorf("expected 0 when any dimension is 0, got %v", r)
	}
	if r := CalculateRisk(5, 5, -1); r != 0 {
		t.Errorf("expected 0 for negative input, got %v", r)
	}
}

func TestCalculateDimension_ScaledGeomeanIdempotent(t *testing.T) {
	// Equal inputs must round-trip unchanged
	if got := CalculateDimension(5, 5); !approx(got, 5.0, 0.1) {
		t.Errorf("scaled geomean(5,5) = %v, want 5", got)
	}
	if got := CalculateDimension(7, 7); !approx(got, 7.0, 0.1) {
		t.Errorf("scaled geomean(7,7) = %v, want 7", got)
	}
}

func TestCalculateDimension_PDFExample(t *testing.T) {
	// PDF doc string example: Natural=6.5, Human=3.2 → ≈ 5.1
	got := CalculateDimension(6.5, 3.2)
	if !approx(got, 5.1, 0.1) {
		t.Errorf("CalculateDimension(6.5, 3.2) = %v, want ≈5.1", got)
	}
}

func TestCalculateDimensionMultiple_MatchesPair(t *testing.T) {
	if a, b := CalculateDimension(4, 7), CalculateDimensionMultiple([]float64{4, 7}); !approx(a, b, 0.05) {
		t.Errorf("Multiple [4,7]=%v vs pair %v should match", b, a)
	}
}

// ============================================================================
// New category aggregators
// ============================================================================

func TestCalculateCategoryGeomean(t *testing.T) {
	// Equal inputs → same value
	if got := CalculateCategoryGeomean([]float64{5, 5, 5}); !approx(got, 5, 0.1) {
		t.Errorf("geomean(5,5,5) = %v, want 5", got)
	}
	// One zero pulls down the result (geometric semantics)
	got := CalculateCategoryGeomean([]float64{10, 0})
	if got <= 5 || got >= 10 {
		t.Errorf("geomean(10,0) = %v, expected between 5 and 10", got)
	}
}

func TestCalculateCategoryMax(t *testing.T) {
	if got := CalculateCategoryMax([]float64{3, 7, 1}); got != 7 {
		t.Errorf("MAX = %v, want 7", got)
	}
	if got := CalculateCategoryMax([]float64{}); got != 0 {
		t.Errorf("empty MAX should be 0, got %v", got)
	}
}

func TestCalculateWeightedCategory_5025(t *testing.T) {
	// PDF Table 11 weighting: Development&Deprivation 0.5, Inequality 0.25, Aid 0.25
	got := CalculateWeightedCategory(
		[]float64{8, 2, 4},
		[]float64{0.5, 0.25, 0.25},
	)
	want := 8*0.5 + 2*0.25 + 4*0.25 // = 5.5
	if !approx(got, want, 0.1) {
		t.Errorf("WMEAN(8,2,4 / .5,.25,.25) = %v, want %v", got, want)
	}
}

// ============================================================================
// Normalization
// ============================================================================

func TestNormalizeMinMax(t *testing.T) {
	// In-range
	if got := NormalizeMinMax(0.5, 0, 1, false); !approx(got, 5, 0.01) {
		t.Errorf("0.5 in [0,1] = %v, want 5", got)
	}
	// Invert
	if got := NormalizeMinMax(0.5, 0, 1, true); !approx(got, 5, 0.01) {
		t.Errorf("0.5 inverted = %v, want 5", got)
	}
	if got := NormalizeMinMax(1, 0, 1, true); !approx(got, 0, 0.01) {
		t.Errorf("1 inverted in [0,1] = %v, want 0 (best)", got)
	}
	// Clamp out-of-range
	if got := NormalizeMinMax(2, 0, 1, false); !approx(got, 10, 0.01) {
		t.Errorf("clamp above max: %v, want 10", got)
	}
}

// ============================================================================
// Classification
// ============================================================================

func TestClassifyRisk_PDFThresholds(t *testing.T) {
	cases := []struct {
		score float64
		want  string
	}{
		{0.5, "Very Low"}, {1.9, "Very Low"},
		{2.0, "Low"}, {3.4, "Low"},
		{3.5, "Medium"}, {4.9, "Medium"},
		{5.0, "High"}, {6.4, "High"},
		{6.5, "Very High"}, {10, "Very High"},
	}
	for _, c := range cases {
		if got := ClassifyRisk(c.score); got != c.want {
			t.Errorf("Classify(%v) = %v, want %v", c.score, got, c.want)
		}
	}
}

func TestClassifyDimension_PerDimensionThresholds(t *testing.T) {
	// Tanzania HAZARD VH = 4.7; a 4.8 should be "Very High" for HAZARD
	// but "Low" for COPING_CAPACITY (LCC VH = 7.7, LCC L cutoff = 5.3)
	if got := ClassifyDimension(4.8, "HAZARD", "TANZANIA"); got != "Very High" {
		t.Errorf("TZ HAZARD(4.8) = %v, want Very High", got)
	}
	if got := ClassifyDimension(4.8, "COPING_CAPACITY", "TANZANIA"); got != "Low" {
		t.Errorf("TZ LCC(4.8) = %v, want Low", got)
	}
	// Global thresholds differ. PDF Table 21 HAZARD: VL<1.4, L<2.6, M<4.0, H<6.0, VH≥6.0
	// 4.8 falls in [4.0, 6.0) → "High"
	if got := ClassifyDimension(4.8, "HAZARD", "GLOBAL"); got != "High" {
		t.Errorf("GLOBAL HAZARD(4.8) = %v, want High", got)
	}
	if got := ClassifyDimension(3.5, "HAZARD", "GLOBAL"); got != "Medium" {
		t.Errorf("GLOBAL HAZARD(3.5) = %v, want Medium", got)
	}
}

// ============================================================================
// Lack of Reliability Index
// ============================================================================

func TestCalculateLRI_FullyAvailable(t *testing.T) {
	lri := CalculateLRI(50, 0, makeGaps(50, 0), false)
	if !approx(lri.Score, 0, 0.01) {
		t.Errorf("LRI with no missing/no staleness = %v, want 0", lri.Score)
	}
	if lri.Classification != "Very Low (most reliable)" {
		t.Errorf("classification = %v", lri.Classification)
	}
}

func TestCalculateLRI_HalfMissingMaxStale(t *testing.T) {
	lri := CalculateLRI(50, 25, makeGaps(25, 10), false)
	if lri.Score < 8 || lri.Score > 10 {
		t.Errorf("LRI half-missing+max-stale = %v, want 8-10", lri.Score)
	}
}

func TestCalculateLRI_ConflictAggravator(t *testing.T) {
	base := CalculateLRI(50, 10, makeGaps(40, 2), false)
	conflict := CalculateLRI(50, 10, makeGaps(40, 2), true)
	expected := math.Min(10, base.Score*1.3)
	if !approx(conflict.Score, expected, 0.05) {
		t.Errorf("conflict LRI = %v, want %v (base %v × 1.3)", conflict.Score, expected, base.Score)
	}
	if !conflict.ConflictAggravator {
		t.Error("ConflictAggravator flag not set")
	}
}

func TestCalculateLRI_ZeroIndicators(t *testing.T) {
	if r := CalculateLRI(0, 0, nil, false); r.Score != 0 {
		t.Errorf("zero indicators should return zero LRI, got %v", r.Score)
	}
}

func makeGaps(n int, val float64) []float64 {
	g := make([]float64, n)
	for i := range g {
		g[i] = val
	}
	return g
}

// ============================================================================
// Property tests
// ============================================================================

func TestScaledGeomean_Monotonic(t *testing.T) {
	for a := 0.0; a <= 9.0; a += 1 {
		for b := 0.0; b <= 9.0; b += 1 {
			before := CalculateDimension(a, b)
			after := CalculateDimension(a+1, b)
			if after < before-0.05 {
				t.Errorf("scaled geomean not monotonic: f(%v,%v)=%v vs f(%v,%v)=%v",
					a, b, before, a+1, b, after)
			}
		}
	}
}

func TestRiskScore_BoundedTo0_10(t *testing.T) {
	// Sample a 4D grid
	for h := 0.0; h <= 10; h += 2 {
		for v := 0.0; v <= 10; v += 2 {
			for lcc := 0.0; lcc <= 10; lcc += 2 {
				r := CalculateRisk(h, v, lcc)
				if r < 0 || r > 10 {
					t.Errorf("Risk(%v,%v,%v) = %v outside [0,10]", h, v, lcc, r)
				}
			}
		}
	}
}
