package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

func LogLevelFromString(s string) zerolog.Level {
	switch strings.ToLower(s) {
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "debug":
		return zerolog.DebugLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	}

	return zerolog.Disabled
}

type Logger struct {
	Zerolog zerolog.Logger
}

func NewLogger(level string, destination io.Writer) *Logger {
	zerologLevel := LogLevelFromString(level)

	if destination == nil {
		destination = os.Stdin
	}

	return &Logger{zerolog.New(destination).Level(zerologLevel).With().Timestamp().Logger()}
}

func (l *Logger) SetPrefix(prefix string) {
	l.Zerolog = l.Zerolog.With().Str("prefix", prefix).Logger()
}

var EmptyLogger = NewLogger("NONE", nil)

func (l *Logger) Info(v ...any)  { l.Zerolog.Info().Msg(fmt.Sprint(v...)) }
func (l *Logger) Warn(v ...any)  { l.Zerolog.Warn().Msg(fmt.Sprint(v...)) }
func (l *Logger) Debug(v ...any) { l.Zerolog.Debug().Msg(fmt.Sprint(v...)) }
func (l *Logger) Error(v ...any) { l.Zerolog.Error().Msg(fmt.Sprint(v...)) }
func (l *Logger) Fatal(v ...any) { l.Zerolog.Fatal().Msg(fmt.Sprint(v...)); os.Exit(1) }

func (l *Logger) Infof(format string, v ...any)  { l.Zerolog.Info().Msgf(format, v...) }
func (l *Logger) Warnf(format string, v ...any)  { l.Zerolog.Warn().Msgf(format, v...) }
func (l *Logger) Debugf(format string, v ...any) { l.Zerolog.Debug().Msgf(format, v...) }
func (l *Logger) Errorf(format string, v ...any) { l.Zerolog.Error().Msgf(format, v...) }
func (l *Logger) Fatalf(format string, v ...any) { l.Zerolog.Fatal().Msgf(format, v...); os.Exit(1) }
