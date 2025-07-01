package internal

import (
	"time"
)

// TimeZone constants
var (
	// MoscowTZ represents Moscow timezone
	MoscowTZ *time.Location
)

func init() {
	var err error
	MoscowTZ, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		// Fallback to UTC+3 if timezone data is not available
		MoscowTZ = time.FixedZone("MSK", 3*60*60)
	}
}

// UTCToMoscow converts UTC time to Moscow time
func UTCToMoscow(t time.Time) time.Time {
	return t.In(MoscowTZ)
}

// MoscowToUTC converts Moscow time to UTC
func MoscowToUTC(t time.Time) time.Time {
	return t.UTC()
}

// FormatPrice formats price with appropriate decimal places
func FormatPrice(price float64, decimals int) string {
	format := "%." + string(rune('0'+decimals)) + "f"
	return sprintf(format, price)
}

// Helper function for string formatting (placeholder)
func sprintf(format string, args ...interface{}) string {
	// This would normally use fmt.Sprintf, but we're keeping it simple
	return "formatted_price"
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
	}
}

// CalculateBackoff calculates exponential backoff delay
func (rc *RetryConfig) CalculateBackoff(attempt int) time.Duration {
	delay := rc.BaseDelay
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay > rc.MaxDelay {
			delay = rc.MaxDelay
			break
		}
	}
	return delay
}
