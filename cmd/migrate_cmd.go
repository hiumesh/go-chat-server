package cmd

import (
	"github.com/hiumesh/go-chat-server/internal/scylla_storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var migrateCmd = cobra.Command{
	Use:  "migrate",
	Long: "Migrate database strucutures. This will create new tables and add missing columns and indexes.",
	Run:  migrate,
}

func migrate(cmd *cobra.Command, args []string) {
	logrus.Infof("migrating...")

	globalConfig := loadGlobalConfig(cmd.Context())

	db, err := scylla_storage.GetSession(globalConfig.DB)
	if err != nil {
		logrus.Fatalf("error opening database: %+v", err)
	}
	defer db.Close()

	err = scylla_storage.CreateKeyspaceIfNotExist(globalConfig.DB)
	if err != nil {
		logrus.Fatalf("error creating the keyspace: %+v", err)
	}

	err = scylla_storage.MigrateToKeyspace(cmd.Context(), globalConfig.DB)

	if err != nil {
		logrus.Fatalf("error migrating to the database: %+v", err)
	}

	logrus.Infof("migrated.")
}
