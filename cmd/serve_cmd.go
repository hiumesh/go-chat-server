package cmd

import (
	"net"

	"github.com/hiumesh/go-chat-server/internal/api"
	"github.com/hiumesh/go-chat-server/internal/conf"
	"github.com/hiumesh/go-chat-server/internal/redis_storage"
	"github.com/hiumesh/go-chat-server/internal/scylla_storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serveCmd = cobra.Command{
	Use:  "serve",
	Long: "Start Web Socket Server",
	Run: func(cmd *cobra.Command, args []string) {
		serve(cmd)
	},
}

func serve(cmd *cobra.Command) {
	globalConfig, err := conf.LoadGlobal("")
	if err != nil {
		logrus.WithError(err).Fatal("unable to load config")
	}

	logrus.Info(globalConfig.REDIS.URL)

	db, err := scylla_storage.Dial(&globalConfig.DB)
	if err != nil {
		logrus.Fatalf("error opening scylla database: %+v", err)
	}
	defer db.Close()

	redisDb, err := redis_storage.Dial(&globalConfig.REDIS)
	if err != nil {
		logrus.Fatalf("error opening redis database: %+v", err)
	}
	defer redisDb.Close()

	api := api.NewAPIWithVersion(cmd.Context(), globalConfig, db, redisDb, "latest")

	addr := net.JoinHostPort(globalConfig.API.Host, globalConfig.API.Port)
	logrus.Infof("GoTrue API started on: %s", addr)

	api.ListenAndServe(cmd.Context(), addr)
}
