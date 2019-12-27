package time

import "time"

// Round rounds the given Time t to the Duration unit.
func Round(t time.Time, unit time.Duration) time.Time {
	if unit == 0 {
		return t
	}
	u := int64(unit)
	diff := t.UnixNano() % u
	rounded := t.UnixNano() - diff
	if diff*2 >= u {
		rounded += u
	}
	return time.Unix(0, rounded)
}
