package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hiumesh/go-chat-server/internal/conf"
	"github.com/hiumesh/go-chat-server/internal/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/scylladb/gocqlx/v2"
	"github.com/sirupsen/logrus"
)

const (
	defaultVersion = "default version"
)

var bearerRegexp = regexp.MustCompile(`^(?:B|b)earer (\S+$)`)

type API struct {
	handler *gin.Engine
	db      gocqlx.Session
	config  *conf.GlobalConfiguration
	version string
}

func NewAPI(globalConfig *conf.GlobalConfiguration, db gocqlx.Session, redisDb *redis.Client) *API {
	return NewAPIWithVersion(context.Background(), globalConfig, db, redisDb, defaultVersion)
}

func NewAPIWithVersion(ctx context.Context, globalConfig *conf.GlobalConfiguration, db gocqlx.Session, redisDb *redis.Client, version string) *API {
	api := API{config: globalConfig, db: db, version: version}

	router := gin.Default()

	corsHandler := cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     globalConfig.CORS.AllAllowedHeaders([]string{"Accept", "Authorization", "Content-Type", "X-Client-IP", "X-Client-Info"}),
		ExposeHeaders:    []string{"X-Total-Count", "Link"},
		AllowCredentials: true,
	})
	router.Use(corsHandler)

	manager := websocket.NewManager(ctx, globalConfig, redisDb, db)

	router.Use(addUniqueRequestID(globalConfig))

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "ping"})
	})

	router.Use(api.requireAuthentication).GET("/ws", func(ginCtx *gin.Context) {
		manager.ServeWS(ginCtx)
	})

	api.handler = router
	return &api
}

func (a *API) ListenAndServe(ctx context.Context, hostAndPort string) {
	baseCtx, cancel := context.WithCancel(context.Background())

	log := logrus.WithField("component", "api")

	server := &http.Server{
		Addr:              hostAndPort,
		Handler:           a.handler,
		ReadHeaderTimeout: 2 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return baseCtx
		},
	}

	cleanupWaitGroup.Add(1)
	go func() {
		defer cleanupWaitGroup.Done()

		<-ctx.Done()

		defer cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Minute)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
			log.WithError(err).Error("shutdown failed")
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.WithError(err).Fatal("http server listen failed")
	}
}
