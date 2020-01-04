package average

import (
	"sync"
	"time"

	"github.com/theMomax/openefs/cache/generic"
	"github.com/theMomax/openefs/config"
	models "github.com/theMomax/openefs/models/production"
	"github.com/theMomax/openefs/utils/numbers"
	timeutils "github.com/theMomax/openefs/utils/time"
)

// Config paths
const (
	PathHalfLife = "cache.production.average.halflife"
)

func init() {
	config.RootCtx.PersistentFlags().Float64(PathHalfLife, 720, "the amount of updates after which a single value looses half its weight in the production-model's average-day recording")
	config.Viper.BindPFlag(PathHalfLife, config.RootCtx.PersistentFlags().Lookup(PathHalfLife))

	config.OnInitialize(func() {
		halfLife = config.Viper.GetFloat64(PathHalfLife)
		outdatedAfter = config.Viper.GetDuration(models.PathStepSize)
		cache = generic.NewCache(outdated)
	})
}

type element struct {
	derived    *numbers.Average
	nonderived *numbers.Average
	m          *sync.Mutex
	daysAhead  uint
	hourOfDay  uint
}

var halfLife float64

var outdatedAfter time.Duration

var cache *generic.Cache

// Run initializes the caching package.
func Run() {
	models.Subscribe(func(u models.Update) {
		dist := models.Round(u.Time()).Sub(models.Round(timeutils.Now()))
		daysAhead := uint(dist.Truncate(24*time.Hour) / (24 * time.Hour))
		hourOfDay := uint(u.Time().Hour()) // uint((dist - (time.Duration(daysAhead) * 24 * time.Hour)) / time.Hour)

		v, ok := cache.Get(daysAhead*24 + hourOfDay).(*element)
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
		cache.Update(v)
	})
}

func outdated(at interface{}) bool {
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
	v, ok := cache.Get(daysAhead*24 + hourOfDay).(*element)
	if !ok {
		return 0.0, false
	}
	v.m.Lock()
	defer v.m.Unlock()
	return v.derived.Get(), true
}

// GetNonDerived returns the average non-derived power for time t.
func GetNonDerived(daysAhead, hourOfDay uint) (val float64, ok bool) {
	v, ok := cache.Get(daysAhead*24 + hourOfDay).(*element)
	if !ok {
		return 0.0, false
	}
	v.m.Lock()
	defer v.m.Unlock()
	return v.nonderived.Get(), true
}
