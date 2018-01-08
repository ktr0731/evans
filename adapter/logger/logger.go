package logger

import (
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/usecase/port"
)

func NewStdLogger(config *config.Config) port.Logger {
	return log.New(os.Stdout, config.Log.Prefix, log.LstdFlags)
}

func NewPromptLogger(ui io.Writer, config *config.Config) port.Logger {
	return log.New(ui, config.Log.Prefix, log.LstdFlags)
}

func NewNopLogger() port.Logger {
	return log.New(ioutil.Discard, "", log.LstdFlags)
}
