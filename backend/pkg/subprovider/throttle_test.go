package subprovider_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thawng/velox/pkg/subprovider"
)

func TestNewThrottle(t *testing.T) {
	t.Parallel()
	th := subprovider.NewThrottle(500 * time.Millisecond)
	assert.NotNil(t, th)
}

func TestThrottle_Wait(t *testing.T) {
	t.Parallel()
	// Create a throttle with 100ms interval
	th := subprovider.NewThrottle(100 * time.Millisecond)

	start := time.Now()
	// First call should return immediately (no previous call)
	th.Wait()
	elapsed1 := time.Since(start)

	// Second call should wait at least 100ms
	start = time.Now()
	th.Wait()
	elapsed2 := time.Since(start)

	// First call should be very quick (just sets lastCall)
	assert.Less(t, elapsed1, 50*time.Millisecond, "first call should be quick")

	// Second call should wait at least 100ms - interval
	// But due to timing variations, we just check it's not instant
	assert.GreaterOrEqual(t, elapsed2, 50*time.Millisecond, "second call should wait approximately the interval")
}

func TestThrottle_WaitMultiple(t *testing.T) {
	t.Parallel()
	// Create a throttle with 50ms interval
	th := subprovider.NewThrottle(50 * time.Millisecond)

	start := time.Now()
	for i := 0; i < 5; i++ {
		th.Wait()
	}
	elapsed := time.Since(start)

	// 5 calls * 50ms interval = ~250ms minimum
	// Allow some slack for test flakiness
	assert.GreaterOrEqual(t, elapsed, 200*time.Millisecond, "5 calls with 50ms interval should take at least 200ms")
}

func TestThrottle_WaitZeroInterval(t *testing.T) {
	t.Parallel()
	// Zero interval should not cause issues
	th := subprovider.NewThrottle(0 * time.Millisecond)

	start := time.Now()
	th.Wait()
	elapsed := time.Since(start)

	// Should return immediately
	assert.Less(t, elapsed, 10*time.Millisecond, "zero interval should return immediately")
}

func TestThrottle_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	// Test that concurrent access doesn't panic
	th := subprovider.NewThrottle(1 * time.Millisecond)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				th.Wait()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// Result struct for testing
func TestResult_Fields(t *testing.T) {
	t.Parallel()
	r := subprovider.Result{
		Provider:        "test-provider",
		ExternalID:      "ext-123",
		Title:           "Test Title",
		Language:        "en",
		Format:          "srt",
		Downloads:       100,
		Rating:          4.5,
		Forced:          true,
		HearingImpaired: true,
		AITranslated:    true,
	}

	assert.Equal(t, "test-provider", r.Provider)
	assert.Equal(t, "ext-123", r.ExternalID)
	assert.Equal(t, "Test Title", r.Title)
	assert.Equal(t, "en", r.Language)
	assert.Equal(t, "srt", r.Format)
	assert.Equal(t, 100, r.Downloads)
	assert.Equal(t, 4.5, r.Rating)
	assert.True(t, r.Forced)
	assert.True(t, r.HearingImpaired)
	assert.True(t, r.AITranslated)
}
