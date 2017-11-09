package log

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/intel/rmd/util/log/config"
)

// Init does log config init
// TODO : Need to support Model name and file line fields.
func Init() error {
	config := config.NewConfig()
	l, _ := logrus.ParseLevel(config.Level)
	// FIXME (shaohe), we do not support both stdout and file at the same time.
	if config.Stdout {
		logrus.SetOutput(os.Stdout)
	} else {
		f, err := os.OpenFile(
			config.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0660)
		if err != nil {
			return err
		}
		logrus.SetOutput(f)
	}

	logrus.SetLevel(l)
	if config.Env == "production" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	return nil
}
