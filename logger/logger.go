package logger

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/wholesome-ghoul/persona-prototype-6/constants"
	"gopkg.in/natefinch/lumberjack.v2"
)

var once sync.Once
var log Logger

type Logger struct {
	zerolog.Logger
}

type Event struct {
	zerolog.Event
}

func (e *Event) Msgf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

func Log() Logger {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel, err := strconv.Atoi(os.Getenv(constants.LOG_LEVEL))
		if err != nil {
			logLevel = int(zerolog.InfoLevel) // default to INFO
		}

		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		if os.Getenv(constants.APP_ENV) != constants.DEV {
			fileLogger := &lumberjack.Logger{
				Filename:   "./logs/pp7.log",
				MaxSize:    5,
				MaxBackups: 10,
				MaxAge:     7,
				Compress:   true,
			}

			output = zerolog.MultiLevelWriter(os.Stderr, fileLogger)
		}

		// buildInfo, _ := debug.ReadBuildInfo()
		log = Logger{
			zerolog.New(output).
				Level(zerolog.Level(logLevel)).
				Sample(zerolog.LevelSampler{
					DebugSampler: &zerolog.BasicSampler{N: 10},
				}).
				With().
				Timestamp().
				Caller().
				// Str("go_version", buildInfo.GoVersion).
				Logger(),
		}
	})

	return log
}
