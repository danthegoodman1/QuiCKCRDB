package quickcrdb

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func init() {
	logger = NewLogger()
	zerolog.DefaultContextLogger = &logger
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		function := ""
		fun := runtime.FuncForPC(pc)
		if fun != nil {
			funName := fun.Name()
			slash := strings.LastIndex(funName, "/")
			if slash > 0 {
				funName = funName[slash+1:]
			}
			function = " " + funName + "()"
		}
		return file + ":" + strconv.Itoa(line) + function
	}
}

func NewLogger() zerolog.Logger {
	if os.Getenv("LOG_TIME_MS") == "1" {
		// Log with milliseconds
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	} else {
		zerolog.TimeFieldFormat = time.RFC3339Nano
	}

	zerolog.LevelFieldName = "level"

	zerolog.TimestampFieldName = "time"

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger = logger.Hook(CallerHook{})
	if os.Getenv("QUICK_LOG_DISABLED") == "1" {
		zerolog.SetGlobalLevel(zerolog.NoLevel)
	} else if os.Getenv("QUICK_DEBUG") == "1" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else if os.Getenv("QUICK_INFO") == "1" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	return logger
}

type CallerHook struct{}

func (h CallerHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	e.Caller(3)
}
