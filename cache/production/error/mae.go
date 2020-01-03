package error

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/theMomax/openefs/cache/generic"
	"github.com/theMomax/openefs/config"
	models "github.com/theMomax/openefs/models/production"
	"github.com/theMomax/openefs/utils/numbers"
	timeutils "github.com/theMomax/openefs/utils/time"
)

// Config paths
const (
	PathHalfLife = "cache.production.error.halflife"
)

func init() {
	config.RootCtx.PersistentFlags().Float64(PathHalfLife, 720, "the amount of updates after which a single value looses half its weight in the production-model's error score")
	config.Viper.BindPFlag(PathHalfLife, config.RootCtx.PersistentFlags().Lookup(PathHalfLife))

	config.OnInitialize(func() {
		halfLife = config.Viper.GetFloat64(PathHalfLife)
		outdatedAfter = config.Viper.GetDuration(models.PathStepSize)
		cache = generic.NewCache(outdated)
	})

	config.OnInitialize(func() {
		log = config.NewLogger()
	})
}

var log *logrus.Logger

type element struct {
	predictions map[time.Duration]models.Data
	date        time.Time
}

var (
	outdatedAfter time.Duration
	halfLife      float64
)

var cache *generic.Cache

var emap = make(map[time.Duration]*numbers.Average)
var emapm = &sync.RWMutex{}

var completed time.Time
var completedm = &sync.RWMutex{}

// Run initializes the caching package.
func Run() {
	models.Subscribe(func(u models.Update) {
		// if actual value is not known yet, cache predicted ones
		if u.IsDerived() {
			log.WithField("time", u.Time()).WithField("value", u.Data().Power).Trace("errorcache received derived update")
			e := get(u.Time())
			log.Trace(e)
			if e == nil {
				log.Trace("initialized e")
				e = &element{
					predictions: make(map[time.Duration]models.Data),
					date:        u.Time(),
				}
			}

			log.Trace("set prediction at duration ", models.Round(e.date).Sub(models.Round(timeutils.Now())).String())
			e.predictions[models.Round(e.date).Sub(models.Round(timeutils.Now()))] = *u.Data()

			cache.Update(e)
			return
		}

		log.WithField("time", u.Time()).WithField("value", u.Data().Power).Trace("errorcache received original update")
		// otherwise calculate error
		if e := get(u.Time()); e != nil {
			log.Trace(e)
			log.Trace(e.predictions)
			emapm.Lock()
			defer emapm.Unlock()
			for d, v := range e.predictions {
				if emap[d] == nil {
					log.Trace("initialized MAE for ", d.String(), " ahead")
					emap[d] = numbers.NewMAE(halfLife)
				}
				emap[d].Apply(u.Data().Power, v.Power)
				log.WithField("value", emap[d].Get()).WithField("duration_ahead", d.String()).Info("updated production-error")
			}
		}
		completedm.Lock()
		if completed.Sub(u.Time()) < 0 {
			log.Trace("updated completed from ", completed.String(), " to ", u.Time().String())
			completed = u.Time()
		}
		completedm.Unlock()
	})

}

func outdated(hash interface{}) bool {
	t, ok := hash.(time.Time)
	completedm.RLock()
	defer completedm.RUnlock()
	b := !ok || completed.Sub(t) >= outdatedAfter
	if b {
		log.Trace(t, " outdated (", completed, ") !!!")
	}
	return b
}

func (e *element) Time() time.Time {
	return e.date
}

func (e *element) Hash() interface{} {
	return models.Round(e.Time())
}

// MAE returns the production-model's mean absolute error, where d is the
// duration between realtime and the point in time, where the model predicted
// the values.
func MAE(d time.Duration) (val float64, ok bool) {
	emapm.RLock()
	defer emapm.RUnlock()
	if a := emap[d]; a != nil {
		return a.Get(), true
	}
	return 0, false
}

func get(t time.Time) *element {
	v, ok := cache.Get(models.Round(t)).(*element)
	if !ok {
		return nil
	}
	return v
}
