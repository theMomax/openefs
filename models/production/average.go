package production

import (
	"sync"
	"time"

	"github.com/theMomax/openefs/cache/generic"
	"github.com/theMomax/openefs/config"
	"github.com/theMomax/openefs/utils/numbers"
	timeutils "github.com/theMomax/openefs/utils/time"
)

// Config paths
const (
	PathHalfLife = "models.production.average.halflife"
)

func init() {
	config.RootCtx.PersistentFlags().Float64(PathHalfLife, 720, "the amount of updates after which a single value looses half its weight in the production-model's average-day recording")
	config.Viper.BindPFlag(PathHalfLife, config.RootCtx.PersistentFlags().Lookup(PathHalfLife))

	config.OnInitialize(func() {
		halfLife = config.Viper.GetFloat64(PathHalfLife)
		outdatedAfter = config.Viper.GetDuration(PathStepSize)
		avgcache = generic.NewCache(avgoutdated)
	})
}

type element struct {
	derived    *numbers.Average
	nonderived *numbers.Average
	m          *sync.Mutex
	daysAhead  uint
	hourOfDay  uint
}

// TODO: rewrite this into an instance that can be used by cache and this package

var halfLife float64

var outdatedAfter time.Duration

var avgcache *generic.Cache

// RunAverage initializes the caching package.
func RunAverage() {
	Subscribe(func(u Update) {
		dist := Round(u.Time()).Sub(Round(timeutils.Now()))
		daysAhead := uint(dist.Truncate(24*time.Hour) / (24 * time.Hour))
		hourOfDay := uint(u.Time().Hour())

		v, ok := avgcache.Get(daysAhead*24 + hourOfDay).(*element)
		if !ok {
			v = &element{
				derived:    numbers.NewAverageSum(halfLife),
				nonderived: numbers.NewAverageSum(halfLife),
				m:          &sync.Mutex{},
				daysAhead:  daysAhead,
				hourOfDay:  hourOfDay,
			}
		}
		v.m.Lock()
		defer v.m.Unlock()
		if u.IsDerived() {
			v.derived.Apply(u.Data().Power)
		} else {
			v.nonderived.Apply(u.Data().Power)
		}
		avgcache.Update(v)
	})
}

func avgoutdated(at interface{}) bool {
	return false
}

func (e *element) Time() time.Time {
	return timeutils.Now().Add(time.Duration(e.daysAhead) * 24 * time.Hour).Add(time.Duration(e.hourOfDay) * time.Hour)
}

func (e *element) Hash() interface{} {
	return e.daysAhead*24 + e.hourOfDay
}

// GetDerived returns the average derived power for time t.
func GetDerived(daysAhead, hourOfDay uint) (val float64, ok bool) {
	v, ok := avgcache.Get(daysAhead*24 + hourOfDay).(*element)
	if !ok {
		return 0.0, false
	}
	v.m.Lock()
	defer v.m.Unlock()
	return v.derived.Get(), true
}

// GetNonDerived returns the average non-derived power for time t.
func GetNonDerived(daysAhead, hourOfDay uint) (val float64, ok bool) {
	v, ok := avgcache.Get(daysAhead*24 + hourOfDay).(*element)
	if !ok {
		return 0.0, false
	}
	v.m.Lock()
	defer v.m.Unlock()
	return v.nonderived.Get(), true
}
