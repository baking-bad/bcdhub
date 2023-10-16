package logger

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Loggable -
type Loggable interface {
	LogFields() map[string]interface{}
}

func New(level string) {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}

	switch os.Getenv("BCD_ENV") {
	case "development":
		consoleWriter.TimeFormat = "2006-01-02 15:04:05"
	default:
		consoleWriter.FormatTimestamp = func(i interface{}) string {
			return ""
		}
	}

	if err := setLevel(level); err != nil {
		log.Err(err).Msg("init log level")
	}

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}
	log.Logger = log.Logger.With().Caller().Logger().Output(consoleWriter)
}

func setLevel(level string) error {
	if level == "" {
		level = "info"
	}
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLevel)
	return nil
}

// Question -
func Question(format string, v ...interface{}) {
	blue := color.New(color.FgMagenta).SprintFunc()
	fmt.Printf("[%s] %s", blue("?"), fmt.Sprintf(format, v...))
}
