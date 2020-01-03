package production

import (
	"time"

	"github.com/theMomax/openefs/cache/generic"
	"github.com/theMomax/openefs/config"
	models "github.com/theMomax/openefs/models/production"
	timeutils "github.com/theMomax/openefs/utils/time"
)

func init() {
	config.OnInitialize(func() {
		outdatedAfter = config.Viper.GetDuration(models.PathStepSize)
		cache = generic.NewCache(outdated)
	})
}

type element struct {
	u models.Update
}

var outdatedAfter time.Duration

var cache *generic.Cache

// Run initializes the caching package.
func Run() {
	models.Subscribe(func(u models.Update) {
		// don't overwrite non-derived value
		if prev := Update(u.Time()); prev == nil || prev.IsDerived() {
			cache.Update(&element{u})
		}
	})
}

func outdated(hash interface{}) bool {
	t, ok := hash.(time.Time)
	return !ok || timeutils.Since(t) >= outdatedAfter
}

func (e *element) Time() time.Time {
	return e.u.Time()
}

func (e *element) Hash() interface{} {
	return models.Round(e.Time())
}

// Update returns the latest available update for time t.
func Update(t time.Time) models.Update {
	v, ok := cache.Get(models.Round(t)).(*element)
	if !ok {
		return nil
	}
	return v.u
}

// Subscribe registers a callback to be called each time, when the underlying
// model creates new output and immediately with the currently cached value. If
// there are (relative) timestamps given, the callback is only called, if the
// update is related to one of those timestamps. It returns the id required for
// unsubscribing. It returns -1, if callback is nil.
func Subscribe(callback func(models.Update), absolute []time.Time, relative []time.Duration) int64 {
	observedHashes := make([]interface{}, len(absolute))
	for i := range absolute {
		observedHashes[i] = interface{}(absolute[i])
	}

	observers := make([]func(interface{}) bool, len(relative))
	for i := range relative {
		observers[i] = func(d time.Duration) func(interface{}) bool {
			return func(hash interface{}) bool {
				return models.Round(timeutils.Now().Add(d)) == hash
			}
		}(relative[i])
	}

	return cache.Subscribe(func(e generic.Element) {
		v, ok := e.(*element)
		if !ok {
			callback(nil)
		}
		callback(v.u)
	}, observedHashes, observers)
}

// Unsubscribe the callback with the given id.
func Unsubscribe(id int64) {
	cache.Unsubscribe(id)
}
