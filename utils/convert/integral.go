package convert

import (
	"errors"
	"time"
)

// Accuracy of the integration-steps
const Accuracy = 5 * time.Minute

// Error constants
var (
	ErrIllegalTimestamps = errors.New("the given timestamps are invalid")
	ErrNoData            = errors.New("could not get power for some point in time")
)

// Integrate the power-function from one point in time to another. If no
// accuracy is specified, Accuracy is used. The unit of power is to be Watts.
// Integrate's unit is kWh.
func Integrate(from time.Time, to time.Time, power func(at time.Time) *float64, accuracy ...time.Duration) (float64, error) {
	if from.Sub(to) > 0 {
		return 0, ErrIllegalTimestamps
	}

	a := Accuracy
	if len(accuracy) == 1 {
		a = accuracy[0]
	}

	sum := 0.0
	count := 0
	for i := from; i.Sub(to) <= 0; i.Add(a) {
		p := power(i)
		if p == nil {
			return 0.0, ErrNoData
		}
		sum += *p
		count++
	}

	avg := sum / float64(count)

	dur := to.Sub(from)

	return (avg / 1000) * dur.Hours(), nil
}
