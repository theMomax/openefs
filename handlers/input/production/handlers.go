package production

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/theMomax/openefs/utils/metadata"

	"github.com/gin-gonic/gin"
	models "github.com/theMomax/openefs/models/production"

	syncutils "github.com/theMomax/openefs/utils/synchronization"
	timeutils "github.com/theMomax/openefs/utils/time"
)

// Register takes care of registering all handler functions to the router.
func Register(r *gin.RouterGroup) {
	g := r.Group("production")
	g.POST("/:unixtimestamp/", handleBasicProductionInput)
}

func handleBasicProductionInput(ctx *gin.Context) {
	unixsecs, err := strconv.ParseInt(ctx.Param("unixtimestamp"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
		return
	}

	timestamp := time.Unix(unixsecs, 0)

	var data models.Data
	ctx.Bind(&data)

	syncutils.AttachID(func(id uint64) {
		if ok := models.UpdateProduction(&update{
			data: &data,
			time: timestamp,
			meta: &metadata.Basic{
				Timestamp:  timeutils.Now(),
				Identifier: id,
			},
		}, 5*time.Second); !ok {
			ctx.AbortWithError(http.StatusIMUsed, errors.New("system is overloaded: model update-pipeline is full"))
		}
	})

	ctx.Status(http.StatusOK)
}
