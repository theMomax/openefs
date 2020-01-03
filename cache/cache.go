package cache

import (
	"github.com/theMomax/openefs/cache/production"
	errorcache "github.com/theMomax/openefs/cache/production/error"
)

// Run initializes the caching package.
func Run() {
	production.Run()
	errorcache.Run()
}
