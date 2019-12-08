package output

import (
	"github.com/gin-gonic/gin"
	"github.com/theMomax/openefs/handlers/output/production"
)

// Register takes care of registering all handler functions to the router.
func Register(r *gin.RouterGroup) {
	g := r.Group("output")
	production.Register(g)
}
