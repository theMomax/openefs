package production

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/theMomax/openefs/config"

	cache "github.com/theMomax/openefs/cache/production"
	avgcache "github.com/theMomax/openefs/cache/production/average"
	errorcache "github.com/theMomax/openefs/cache/production/error"
	models "github.com/theMomax/openefs/models/production"
	"github.com/theMomax/openefs/utils/convert"
	timeutils "github.com/theMomax/openefs/utils/time"

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
	g.GET("/day/relative/:at", handleProductionRequestAtDayRelative)
	g.GET("/day/absolute/:at", handleProductionRequestAtDay)
	g.GET("/day/avg/derived/relative/:at", func(ctx *gin.Context) {
		handleProductionRequestAtDayAvgRelative(ctx, true)
	})
	g.GET("/day/avg/nonderived/relative/:at", func(ctx *gin.Context) {
		handleProductionRequestAtDayAvgRelative(ctx, false)
	})
	g.GET("/day/avg/derived/absolute/:at", func(ctx *gin.Context) {
		handleProductionRequestAtDayAvg(ctx, true)
	})
	g.GET("/day/avg/nonderived/absolute/:at", func(ctx *gin.Context) {
		handleProductionRequestAtDayAvg(ctx, false)
	})
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

func handleProductionRequestAtDay(ctx *gin.Context) {
	atunixsecs, err := strconv.ParseInt(ctx.Param("at"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
		return
	}

	at := time.Unix(atunixsecs, 0)
	start := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, at.Location())
	values := make([]*float64, 24)
	for i := 0; i < 24; i++ {
		if u := cache.Update(start.Add(time.Duration(i) * time.Hour)); u != nil && u.Data() != nil {
			values[i] = &u.Data().Power
		}
	}
	ctx.JSON(http.StatusOK, values)
}

func handleProductionRequestAtDayRelative(ctx *gin.Context) {
	atdays, err := strconv.ParseInt(ctx.Param("at"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
		return
	}

	at := timeutils.Now().Add(24 * time.Hour * time.Duration(atdays))
	start := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, at.Location())
	values := make([]*float64, 24)
	for i := 0; i < 24; i++ {
		if u := cache.Update(start.Add(time.Duration(i) * time.Hour)); u != nil && u.Data() != nil {
			values[i] = &u.Data().Power
		}
	}
	ctx.JSON(http.StatusOK, values)
}

func handleProductionRequestAtDayAvg(ctx *gin.Context, derived bool) {
	atunixsecs, err := strconv.ParseInt(ctx.Param("at"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
		return
	}

	at := time.Unix(atunixsecs, 0)
	start := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, at.Location())
	values := make([]*float64, 24)
	get := avgcache.GetNonDerived
	if derived {
		get = avgcache.GetDerived
	}

	for i := 0; i < 24; i++ {
		if v, ok := get(uint(start.Sub(timeutils.Now()).Truncate(24*time.Hour)/(24*time.Hour)), uint(i)); ok {
			values[i] = &v
		}
	}
	ctx.JSON(http.StatusOK, values)
}

func handleProductionRequestAtDayAvgRelative(ctx *gin.Context, derived bool) {
	atdays, err := strconv.ParseInt(ctx.Param("at"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
		return
	}

	values := make([]*float64, 24)
	get := avgcache.GetNonDerived
	if derived {
		get = avgcache.GetDerived
	}

	for i := 0; i < 24; i++ {
		if v, ok := get(uint(atdays), uint(i)); ok {
			values[i] = &v
		}
	}
	ctx.JSON(http.StatusOK, values)
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
