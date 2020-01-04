package time

import "time"

func YearProcess(t time.Time) float64 {
	return float64(t.YearDay()) / 366.0
}

func DayProcess(t time.Time) float64 {
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return float64(t.Sub(midnight)) / (24 * float64(time.Hour))
}
