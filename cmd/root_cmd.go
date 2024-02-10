package cmd

import (
	"context"

	"github.com/hiumesh/go-chat-server/internal/conf"
	"github.com/hiumesh/go-chat-server/internal/observability"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var configFile = ""

var rootCmd = cobra.Command{
	Use: "gosocket",
	Run: func(cmd *cobra.Command, args []string) {
		migrate(cmd, args)
		serve(cmd)
	},
}

func RootCommand() *cobra.Command {
	rootCmd.AddCommand(&serveCmd, &migrateCmd)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "the config file to use")

	return &rootCmd
}

func loadGlobalConfig(ctx context.Context) *conf.GlobalConfiguration {
	if ctx == nil {
		panic("context must not be nil")
	}

	config, err := conf.LoadGlobal(configFile)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %+v", err)
	}

	if err := observability.ConfigureLogging(&config.LOGGING); err != nil {
		logrus.WithError(err).Error("unable to configure logging")
	}

	return config
}
