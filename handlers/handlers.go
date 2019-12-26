package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/theMomax/openefs/handlers/input"
	"github.com/theMomax/openefs/handlers/output"
)

// Register takes care of registering all handler functions to the router.
func Register(r *gin.RouterGroup) {
	g := r.Group("v1")
	input.Register(g)
	output.Register(g)
}
