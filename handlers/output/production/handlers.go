package production

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/theMomax/openefs/config"

	cache "github.com/theMomax/openefs/cache/production"
	errorcache "github.com/theMomax/openefs/cache/production/error"
	models "github.com/theMomax/openefs/models/production"
	"github.com/theMomax/openefs/utils/convert"

	"github.com/gin-gonic/gin"
)

func init() {
	config.OnInitialize(func() {
		stepSize = config.Viper.GetDuration(models.PathStepSize)
	})
}

var (
	stepSize time.Duration
)

// Register takes care of registering all handler functions to the router.
func Register(r *gin.RouterGroup) {
	g := r.Group("production")
	g.GET("/from/:from/to/:to", handleProductionRequest)
	g.GET("/at/:at", handleProductionRequestAtTime)
	g.GET("/error", handleProductionError)
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

func handleProductionError(ctx *gin.Context) {
	errs := make([]float64, 0)
	d := stepSize
	for {
		e, ok := errorcache.MAE(d)
		if !ok {
			break
		}
		errs = append(errs, e)
		d += stepSize
	}
	ctx.JSON(http.StatusOK, errs)
}
