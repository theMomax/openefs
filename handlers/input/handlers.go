package input

import (
	"github.com/gin-gonic/gin"
	"github.com/theMomax/openefs/handlers/input/production"
	"github.com/theMomax/openefs/handlers/input/weather"
)

// Register takes care of registering all handler functions to the router.
func Register(r *gin.RouterGroup) {
	g := r.Group("input")
	production.Register(g)
	weather.Register(g)
}
