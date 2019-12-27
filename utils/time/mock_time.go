package time

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonboulle/clockwork"
	"github.com/sirupsen/logrus"
	"github.com/theMomax/openefs/config"
)

// Config paths
const (
	PathUseMockTime  = "utils.time.usemocktime"
	PathMockTimeIP   = "utils.time.mocktimeip"
	PathMockTimePort = "utils.time.mocktimeport"
)

var clock = clockwork.NewRealClock()
var fakeClock clockwork.FakeClock
var fakeClockReady = &sync.WaitGroup{}

func init() {
	config.RootCtx.PersistentFlags().Bool(PathUseMockTime, false, "flag for enabling mocked-time (the time can be set using http-endpoint)")
	config.Viper.BindPFlag(PathUseMockTime, config.RootCtx.PersistentFlags().Lookup(PathUseMockTime))

	config.RootCtx.PersistentFlags().String(PathMockTimeIP, "localhost", "address for mock-time endpoint")
	config.Viper.BindPFlag(PathMockTimeIP, config.RootCtx.PersistentFlags().Lookup(PathMockTimeIP))

	config.RootCtx.PersistentFlags().Uint(PathMockTimePort, 8090, "port for mock-time endpoint")
	config.Viper.BindPFlag(PathMockTimePort, config.RootCtx.PersistentFlags().Lookup(PathMockTimePort))

	config.OnInitialize(func() {
		log = config.NewLogger()
	})
	config.OnInitialize(func() {
		if config.Viper.GetBool(PathUseMockTime) {
			log.Warning("mock-time enabled")
			fakeClockReady.Add(1)
			go startMockServer()
			log.WithField("address", config.Viper.GetString(PathMockTimeIP)+":"+config.Viper.GetString(PathMockTimePort)+"/utils/time/mocktime/:unixtimestamp").Info("waiting for initial mock-time to be set...")
			fakeClockReady.Wait()
		}
	})
}

var log *logrus.Logger

func Now() time.Time {
	return clock.Now()
}

func Sleep(d time.Duration) {
	clock.Sleep(d)
}

func Since(t time.Time) time.Duration {
	return Now().Sub(t)
}

func After(d time.Duration) <-chan time.Time {
	return clock.After(d)
}

func startMockServer() error {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(gin.Recovery())

	r.Use(config.GinLogrusLogger())

	r.GET("utils/time/mocktime/:unixtimestamp", mockTimeHandler)

	return r.Run(config.Viper.GetString(PathMockTimeIP) + ":" + config.Viper.GetString(PathMockTimePort))
}

func mockTimeHandler(ctx *gin.Context) {
	unixsecs, err := strconv.ParseInt(ctx.Param("unixtimestamp"), 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
		return
	}
	t := time.Unix(unixsecs, 0)

	if fakeClock == nil {
		fakeClock = clockwork.NewFakeClockAt(t)
		clock = fakeClock
		fakeClockReady.Done()
		return
	}

	fakeClock.Advance(t.Sub(Now()))
}
