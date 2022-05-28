package logging

import (
	zerolog "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

func InitAppLogger(logLevel string) *Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	parseLogLvl(logLevel)
	return &Logger{
		logger: zerolog.New(os.Stderr).With().Timestamp().Logger(),
	}
}

type Logger struct {
	logger zerolog.Logger
}

func (l *Logger) EasyLogInfo(prefix, message, data string) {
	l.logger.Info().
		Str("service", prefix).
		Msgf(message + data)
}

func (l *Logger) EasyLogError(prefix, message, data string, reportedErr error) {
	l.logger.Error().
		Err(reportedErr).
		Str("service", prefix).
		Msgf(message + data)
}

func (l *Logger) EasyLogFatal(prefix, message, data string, reportedErr error) {
	log.Fatal().
		Err(reportedErr).
		Str("service", prefix).
		Msgf(message + data)
}

func (l *Logger) EasyLogDebug(prefix, message, data string) {
	log.Debug().
		Str("service", prefix).
		Msgf(message + data)
}

func parseLogLvl(logLevel string) {
	logLevel = strings.ToUpper(logLevel)
	switch {
	case logLevel == "FATAL":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
		return
	case logLevel == "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		return
	case logLevel == "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		return
	case logLevel == "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		return
	case logLevel == "DEBUG":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		return
	default:
		// will return INFO_LEVEL
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		return
	}
}
