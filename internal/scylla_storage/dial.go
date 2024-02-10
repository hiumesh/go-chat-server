package scylla_storage

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gocql/gocql"
	"github.com/hiumesh/go-chat-server/internal/conf"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/migrate"
)

func Dial(config *conf.DBConfiguration) (gocqlx.Session, error) {
	clusterConfig := GetClusterConfig(config)
	clusterConfig.Keyspace = config.Keyspace
	session, err := gocqlx.WrapSession(clusterConfig.CreateSession())
	if err != nil {
		return session, err
	}
	return session, nil
}

func GetClusterConfig(config *conf.DBConfiguration) *gocql.ClusterConfig {
	retryPolicy := &gocql.ExponentialBackoffRetryPolicy{
		Min:        time.Second,
		Max:        10 * time.Second,
		NumRetries: 5,
	}

	cluster := gocql.NewCluster(config.Host)
	cluster.Timeout = 5 * time.Second
	cluster.RetryPolicy = retryPolicy
	// cluster.Keyspace = config.Keyspace
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	cluster.Consistency = gocql.LocalQuorum
	return cluster
}

func GetSession(config conf.DBConfiguration) (gocqlx.Session, error) {
	clusterConfig := GetClusterConfig(&config)
	session, err := gocqlx.WrapSession(clusterConfig.CreateSession())
	if err != nil {
		return session, err
	}
	return session, nil
}

func GetKeyspaceSession(config conf.DBConfiguration) (gocqlx.Session, error) {
	clusterConfig := GetClusterConfig(&config)
	clusterConfig.Keyspace = config.Keyspace
	session, err := gocqlx.WrapSession(clusterConfig.CreateSession())
	if err != nil {
		return session, err
	}
	return session, nil
}

func CreateKeyspaceIfNotExist(config conf.DBConfiguration) error {
	session, err := GetSession(config)

	if err != nil {
		return err
	}

	if err := session.ExecStmt(fmt.Sprintf(
		`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`,
		config.Keyspace,
	)); err != nil {
		return err
	}
	return nil
}

func MigrateToKeyspace(contextx context.Context, config conf.DBConfiguration) error {
	session, err := GetKeyspaceSession(config)

	if err != nil {
		return err
	}

	log := func(ctx context.Context, session gocqlx.Session, ev migrate.CallbackEvent, name string) error {
		return nil
	}

	reg := migrate.CallbackRegister{}
	reg.Add(migrate.BeforeMigration, "before", log)
	reg.Add(migrate.AfterMigration, "after", log)
	reg.Add(migrate.CallComment, "1", log)
	reg.Add(migrate.CallComment, "2", log)
	reg.Add(migrate.CallComment, "3", log)

	migrate.Callback = reg.Callback

	err = migrate.FromFS(contextx, session, os.DirFS(config.MigrationsPath))
	if err != nil {
		return err
	}
	return nil
}
