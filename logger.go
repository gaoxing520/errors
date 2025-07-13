package errors

import (
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog"
)

var (
	// DefaultLogger is the global zerolog logger instance used by this package's logging functions.
	DefaultLogger zerolog.Logger

	// DefaultLogOutput is the default output writer for the logger if not set externally.
	DefaultLogOutput io.Writer

	defaultLoggerOnce sync.Once
	loggerMutex       sync.RWMutex
)

// Trace returns a trace level logger event.
func Trace() *zerolog.Event {
	return getLogger().Trace()
}

// Debug returns a debug level logger event.
func Debug() *zerolog.Event {
	return getLogger().Debug()
}

// Info returns an info level logger event.
func Info() *zerolog.Event {
	return getLogger().Info()
}

// Warn returns a warn level logger event.
func Warn() *zerolog.Event {
	return getLogger().Warn()
}

// Error returns an error level logger event.
func Error() *zerolog.Event {
	return getLogger().Error()
}

// Fatal returns a fatal level logger event, which will exit the application after logging.
func Fatal() *zerolog.Event {
	return getLogger().Fatal()
}

// Panic returns a panic level logger event, which will panic after logging.
func Panic() *zerolog.Event {
	return getLogger().Panic()
}

// SetLogger allows users to set a custom logger
func SetLogger(logger zerolog.Logger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	DefaultLogger = logger
}

// SetLogOutput sets the output writer for the default logger
func SetLogOutput(output io.Writer) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	DefaultLogOutput = output
	// Reset the logger to use new output
	defaultLoggerOnce = sync.Once{}
}

func getLogger() *zerolog.Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	defaultLoggerOnce.Do(func() {
		if DefaultLogOutput == nil {
			DefaultLogOutput = os.Stderr
		}

		DefaultLogger = zerolog.New(DefaultLogOutput).
			With().
			Caller().
			Timestamp().
			Logger().
			Level(zerolog.ErrorLevel)
	})

	return &DefaultLogger
}
