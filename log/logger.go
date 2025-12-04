package log

import (
	"fmt"
	"log"
	"strings"
)

var (
	_ Logger = (*ConsoleLogger)(nil)
	_ Logger = (*AnsiColorLogger)(nil)
)

var (
	DefaultLogger Logger = ConsoleLogger{}
)

type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func LevelFromString(str string) Level {
	str = strings.ToLower(str)
	switch str {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

type ConsoleLogger struct {
	Level Level
}

func (logger ConsoleLogger) Debugf(format string, args ...any) {
	if logger.Level > LevelDebug {
		return
	}
	log.Printf("[DEBUG] "+format, args...)
}

func (logger ConsoleLogger) Infof(format string, args ...any) {
	if logger.Level > LevelInfo {
		return
	}
	log.Printf("[INFO] "+format, args...)
}

func (logger ConsoleLogger) Warnf(format string, args ...any) {
	if logger.Level > LevelWarn {
		return
	}
	log.Printf("[WARN] "+format, args...)
}

func (logger ConsoleLogger) Errorf(format string, args ...any) {
	if logger.Level > LevelError {
		return
	}
	log.Printf("[ERROR] "+format, args...)
}

type AnsiColorLogger struct {
	Level       Level
	ColorOutput bool
}

type ansiColorString interface {
	AnsiColorString() string
}

func (logger AnsiColorLogger) Debugf(format string, args ...any) {
	if logger.Level > LevelDebug {
		return
	}
	format, args = logger.resolveFormatAndArgs("DEBUG", format, args...)
	log.Printf(format, args...)
}

func (logger AnsiColorLogger) Infof(format string, args ...any) {
	if logger.Level > LevelInfo {
		return
	}
	format, args = logger.resolveFormatAndArgs("INFO", format, args...)
	log.Printf(format, args...)
}

func (logger AnsiColorLogger) Warnf(format string, args ...any) {
	if logger.Level > LevelWarn {
		return
	}
	format, args = logger.resolveFormatAndArgs("WARN", format, args...)
	log.Printf(format, args...)
}

func (logger AnsiColorLogger) Errorf(format string, args ...any) {
	if logger.Level > LevelError {
		return
	}
	format, args = logger.resolveFormatAndArgs("ERROR", format, args...)
	log.Printf(format, args...)
}

func (logger AnsiColorLogger) resolveFormatAndArgs(level, format string, args ...any) (string, []any) {
	if !logger.ColorOutput {
		return fmt.Sprintf("[%s] %s", level, format), args
	}
	switch level {
	case "DEBUG":
		format = fmt.Sprintf("[%s] %s", AnsiColorGrey("DEBUG").AnsiColorString(), format)
	case "INFO":
		format = fmt.Sprintf("[%s] %s", AnsiColorWhite("INFO").AnsiColorString(), format)
	case "WARN":
		format = fmt.Sprintf("[%s] %s", AnsiColorYellow("WARN").AnsiColorString(), format)
	case "ERROR":
		format = fmt.Sprintf("[%s] %s", AnsiColorRed("ERROR").AnsiColorString(), format)
	}

	coloredArgs := make([]any, len(args))
	for i, arg := range args {
		if s, ok := arg.(ansiColorString); ok {
			coloredArgs[i] = s.AnsiColorString()
		} else {
			coloredArgs[i] = arg
		}
	}
	return format, coloredArgs
}
