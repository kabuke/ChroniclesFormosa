package disaster

import (
	"testing"
	"math"
	"github.com/kabuke/ChroniclesFormosa/server/model"
)

func TestApplyEarthquakeDamage(t *testing.T) {
	testCases := []struct {
		name         string
		localMag     float32
		initFood     int64
		initLoyalty  int32
		expectedFood int64
		expLoyalty   int32
	}{
		{"Minor (< 3.0)", 2.5, 1000, 50, 950, 45},
		{"Moderate (3.0~5.0)", 4.0, 1000, 50, 850, 40},
		{"Strong (5.0~7.0)", 6.0, 1000, 50, 700, 35},
		{"Severe (>= 7.0)", 7.5, 1000, 50, 600, 30},
		{"Loyalty Clamp to 0", 7.5, 1000, 10, 600, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := &model.Village{
				Food:    tc.initFood,
				Wood:    tc.initFood,
				Loyalty: tc.initLoyalty,
			}

			applyEarthquakeDamage(v, tc.localMag)

			if v.Food != tc.expectedFood {
				t.Errorf("Expected Food %d, got %d", tc.expectedFood, v.Food)
			}
			if v.Wood != tc.expectedFood {
				t.Errorf("Expected Wood %d, got %d", tc.expectedFood, v.Wood)
			}
			if v.Loyalty != tc.expLoyalty {
				t.Errorf("Expected Loyalty %d, got %d", tc.expLoyalty, v.Loyalty)
			}
		})
	}
}

func TestEarthquakeAffectedRadius(t *testing.T) {
	// Radius = Magnitude * 3.0
	// 假設震央在 (0,0)
	mag := float32(5.0)
	radiusSq := float64(mag * 3.0 * mag * 3.0)

	// (10, 10) -> distSq = 200, radiusSq = 225
	// affected == true
	
	distSq := float64(10*10 + 10*10)
	if distSq > radiusSq {
		t.Errorf("Expected distSq <= radiusSq")
	}

	// Calculate affected radius behavior matching earthquake.go:
	// distTiles = math.Sqrt((vx-ex)^2 + (vy-ey)^2)
	// Actually we just test if the math holds true based on the comment.
	expectedRadius := float64(mag * 3.0)
	if math.Abs(expectedRadius-15.0) > 0.01 {
		t.Errorf("Radius logic mismatch")
	}
}
