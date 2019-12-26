package production

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	cache "github.com/theMomax/openefs/cache/production"
	"github.com/theMomax/openefs/utils/convert"

	"github.com/gin-gonic/gin"
)

// Register takes care of registering all handler functions to the router.
func Register(r *gin.RouterGroup) {
	g := r.Group("production")
	g.GET("/from/:from/to/:to", handleProductionRequest)
	g.GET("/at/:at", handleProductionRequestAtTime)
}

func handleProductionRequest(ctx *gin.Context) {
	fromunixsecs, err := strconv.ParseInt(ctx.Param("from"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
		return
	}

	from := time.Unix(fromunixsecs, 0)

	tounixsecs, err := strconv.ParseInt(ctx.Param("to"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
		return
	}

	to := time.Unix(tounixsecs, 0)

	kWh, err := convert.Integrate(from, to, func(at time.Time) *float64 {
		update := cache.Update(at)
		if update == nil {
			return nil
		}
		if update.Data() == nil {
			return nil
		}
		return &update.Data().Power
	})

	if err != nil {
		if errors.Is(err, convert.ErrIllegalTimestamps) {
			ctx.AbortWithError(http.StatusBadRequest, err)
		}
		if errors.Is(err, convert.ErrNoData) {
			ctx.AbortWithError(http.StatusNoContent, err)
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, kWh)
}

func handleProductionRequestAtTime(ctx *gin.Context) {
	atunixsecs, err := strconv.ParseInt(ctx.Param("at"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
		return
	}

	at := time.Unix(atunixsecs, 0)

	update := cache.Update(at)
	if update == nil || update.Data() == nil {
		ctx.AbortWithError(http.StatusNoContent, err)
	}

	ctx.JSON(http.StatusOK, update.Data().Power)
}
