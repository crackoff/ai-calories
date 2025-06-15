package helpers

import (
	"fmt"
	"math"
)

// ZeroBasedRange represents a boundary for a set of numbers.
type ZeroBasedRange struct {
	Max        float64
	Domain     int
	Descending bool
}

// IsDescending returns if the range is descending.
func (r ZeroBasedRange) IsDescending() bool {
	return r.Descending
}

// IsZero returns if the ZeroBasedRange has been set or not.
func (r ZeroBasedRange) IsZero() bool {
	return (r.Max == 0 || math.IsNaN(r.Max))
}

// GetMin gets the min value for the continuous range.
func (r ZeroBasedRange) GetMin() float64 {
	return 0
}

// SetMin sets the min value for the continuous range.
func (r *ZeroBasedRange) SetMin(min float64) {
	// ignoring
}

// GetMax returns the max value for the continuous range.
func (r ZeroBasedRange) GetMax() float64 {
	return r.Max
}

// SetMax sets the max value for the continuous range.
func (r *ZeroBasedRange) SetMax(max float64) {
	r.Max = max
}

// GetDelta returns the difference between the min and max value.
func (r ZeroBasedRange) GetDelta() float64 {
	return r.Max
}

// GetDomain returns the range domain.
func (r ZeroBasedRange) GetDomain() int {
	return r.Domain
}

// SetDomain sets the range domain.
func (r *ZeroBasedRange) SetDomain(domain int) {
	r.Domain = domain
}

// String returns a simple string for the ZeroBasedRange.
func (r ZeroBasedRange) String() string {
	return fmt.Sprintf("ZeroBasedRange [%.2f,%.2f] => %d", 0.0, r.Max, r.Domain)
}

// Translate maps a given value into the ZeroBasedRange space.
func (r ZeroBasedRange) Translate(value float64) int {
	ratio := value / r.GetDelta()

	if r.IsDescending() {
		return r.Domain - int(math.Ceil(ratio*float64(r.Domain)))
	}

	return int(math.Ceil(ratio * float64(r.Domain)))
}
