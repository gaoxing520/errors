package errors

import (
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog"
)

var (
	// DefaultLogger is the global zerolog logger instance used by this package's logging functions.
	DefaultLogger *zerolog.Logger

	// DefaultLogOutput is the default output writer for the logger if not set externally.
	DefaultLogOutput io.Writer = os.Stderr

	// DefaultLogLevel is the default log level for the logger if not set externally.
	DefaultLogLevel = zerolog.InfoLevel

	// loggerMutex protects logger access and configuration
	loggerMutex sync.RWMutex
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
func SetLogger(logger *zerolog.Logger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	DefaultLogger = logger
}

// SetLogOutput sets the output writer for the default logger
func SetLogOutput(output io.Writer) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	DefaultLogOutput = output
	DefaultLogger = nil // Force re-initialization
}

// SetLogLevel sets the log level for the default logger
func SetLogLevel(level zerolog.Level) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	DefaultLogLevel = level
	DefaultLogger = nil // Force re-initialization
}

func getLogger() *zerolog.Logger {
	// Fast path: read lock to check if initialized
	loggerMutex.RLock()
	if DefaultLogger != nil {
		logger := DefaultLogger
		loggerMutex.RUnlock()
		return logger
	}
	loggerMutex.RUnlock()

	// Slow path: write lock to initialize
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	// Double-check after acquiring write lock
	if DefaultLogger != nil {
		return DefaultLogger
	}

	// Initialize logger
	newLogger := zerolog.New(DefaultLogOutput).
		With().
		Caller().
		Timestamp().
		Logger().
		Level(DefaultLogLevel)
	DefaultLogger = &newLogger

	return DefaultLogger
}
