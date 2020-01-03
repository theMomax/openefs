package numbers

import "math"

type Average struct {
	sum      float64
	count    float64
	weight   float64
	operator func(...float64) float64
}

// NewMAE returns a temporarely-weighted Mean Absolute Error. I.e. if a new
// value is applied, it takes halfLife other values until the initial value's
// "weight" is only half of the one, that has just been applied. The Apply
// function takes two arguments: the expected and actual value.
func NewMAE(halfLife float64) *Average {
	return NewAverage(math.Pow(0.5, 1/halfLife), ABSDIFF)
}

func NewAverage(weight float64, operator func(...float64) float64) *Average {
	return &Average{
		sum:      0,
		count:    0,
		weight:   weight,
		operator: operator,
	}
}

func NewAverageFromInitial(initial float64, weight float64, operator func(...float64) float64) *Average {
	return &Average{
		sum:      initial,
		count:    1,
		weight:   weight,
		operator: operator,
	}
}

func (a *Average) Apply(args ...float64) {
	a.sum *= a.weight
	a.count *= a.weight

	a.sum = (a.sum + a.operator(args...))
	a.count++
}

func (a *Average) Get() float64 {
	if a.count == 0.0 {
		return 0.0
	}
	return a.sum / a.count
}

func SUM(args ...float64) float64 {
	s := 0.0
	for _, a := range args {
		s += a
	}
	return s
}

func ABSDIFF(args ...float64) float64 {
	if len(args) == 0 {
		return 0
	}
	d := args[0]
	for i := 1; i < len(args); i++ {
		d -= args[i]
	}
	return math.Abs(d)
}
