package observability

import (
	"os"
	"sync"

	"github.com/hiumesh/go-chat-server/internal/conf"
	"github.com/sirupsen/logrus"
)

var (
	loggingOnce sync.Once
)

func ConfigureLogging(config *conf.LoggingConfig) error {
	var err error

	loggingOnce.Do(func() {
		logrus.SetFormatter(&logrus.JSONFormatter{})

		if config.File != "" {
			f, errOpen := os.OpenFile(config.File, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660) //#nosec G302 -- Log files should be rw-rw-r--
			if errOpen != nil {
				err = errOpen
				return
			}
			logrus.SetOutput(f)
			logrus.Infof("Set output file to %s", config.File)
		}

		if config.Level != "" {
			level, errParse := logrus.ParseLevel(config.Level)
			if err != nil {
				err = errParse
				return
			}
			logrus.SetLevel(level)
			logrus.Debug("Set log level to: " + logrus.GetLevel().String())
		}

		f := logrus.Fields{}
		for k, v := range config.Fields {
			f[k] = v
		}
		logrus.WithFields(f)
	})

	return err
}
