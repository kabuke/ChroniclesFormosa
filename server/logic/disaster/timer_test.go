package disaster

import (
	"testing"
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

func TestDisasterFrequency(t *testing.T) {
	// Reset counters
	earthquakeCount := 0
	typhoonCount := 0

	// Mock the triggers to count them instead of actually triggering
	originalTriggerWarning := triggerWarning

	defer func() {
		triggerWarning = originalTriggerWarning
	}()

	triggerWarning = func(dt pb.DisasterType, msg string) {
		if dt == pb.DisasterType_DISASTER_EARTHQUAKE {
			earthquakeCount++
		} else if dt == pb.DisasterType_DISASTER_TYPHOON {
			typhoonCount++
		}
	}

	// simulate one year: 25920 invocations (if 1 tick per second, wait, evaluateDisaster runs every 1 real-life minute? Or every tick?)
	// In the logic: InGameTime.Year() based on evaluateDisaster being called per tick?
	// It's called when StartTimeEngine's ticker ticks. Let's just call evaluateDisaster() N times.
	// Actually, wait, evaluateDisaster calculates probabilities dynamically...
	// Let's call it 365*24 = 8760 times (if it runs every game hour)
	// We'll call it 25920 times to simulate a period.

	// We run it 10 years to avoid flaky tests due to probability
	InGameTime = time.Date(1600, 1, 2, 1, 0, 0, 0, time.UTC) // Day 2, Hour 1
	for i := 0; i < 259200; i++ {
		month := time.Month((i % 12) + 1) // Cycle through months
		evaluateDisaster(month, "早上")
	}

	if earthquakeCount < 1 {
		t.Errorf("Earthquake count out of expected bounds: %d", earthquakeCount)
	}

	if typhoonCount < 1 {
		t.Errorf("Typhoon count out of expected bounds: %d", typhoonCount)
	}
}
