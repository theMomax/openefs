package cache

import (
	"github.com/theMomax/openefs/cache/production"
)

// Run initializes the caching package.
func Run() {
	production.Run()
}
