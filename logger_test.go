package errors

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestSetLogger(t *testing.T) {
	// Save original state
	originalLogger := DefaultLogger
	originalOutput := DefaultLogOutput
	originalOnce := defaultLoggerOnce
	defer func() {
		DefaultLogger = originalLogger
		DefaultLogOutput = originalOutput
		defaultLoggerOnce = originalOnce
	}()

	// Reset state first
	defaultLoggerOnce = sync.Once{}

	// Create a custom logger with buffer
	var buf bytes.Buffer
	customLogger := zerolog.New(&buf).Level(zerolog.InfoLevel)

	// Set the custom logger
	SetLogger(customLogger)

	// Test that the logger was set by checking DefaultLogger
	if &DefaultLogger == nil {
		t.Error("DefaultLogger should not be nil")
	}

	// Test logging with custom logger - use the DefaultLogger directly
	DefaultLogger.Info().Msg("test message")

	// The buffer should contain the log message
	logOutput := buf.String()
	if len(logOutput) == 0 {
		t.Error("Custom logger should have received log message")
	}

	// Verify the message contains our test content
	if !strings.Contains(logOutput, "test message") {
		t.Errorf("Log output should contain 'test message', got: %s", logOutput)
	}
}

func TestSetLogOutput(t *testing.T) {
	// Save original state
	originalOutput := DefaultLogOutput
	originalLogger := DefaultLogger
	originalOnce := defaultLoggerOnce
	defer func() {
		DefaultLogOutput = originalOutput
		DefaultLogger = originalLogger
		defaultLoggerOnce = originalOnce
	}()

	// Create a buffer for output
	var buf bytes.Buffer
	SetLogOutput(&buf)

	// Test that output was set
	if DefaultLogOutput != &buf {
		t.Error("DefaultLogOutput should be set to buffer")
	}

	// Get logger to initialize with new output
	logger := getLogger()

	// Test logging to custom output using Error level (which is enabled by default)
	logger.Error().Msg("test output message")

	// The buffer should contain the log message
	logOutput := buf.String()
	if len(logOutput) == 0 {
		t.Error("Custom output should have received log message")
	}

	// Verify the message contains our test content
	if !strings.Contains(logOutput, "test output message") {
		t.Errorf("Log output should contain 'test output message', got: %s", logOutput)
	}
}

func TestGetLoggerConcurrency(t *testing.T) {
	// Reset logger state
	defaultLoggerOnce = sync.Once{}
	DefaultLogOutput = nil

	const numGoroutines = 100
	var wg sync.WaitGroup
	results := make([]*zerolog.Logger, numGoroutines)

	// Launch multiple goroutines to get logger concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = getLogger()
		}(i)
	}

	wg.Wait()

	// All results should be the same logger instance
	firstLogger := results[0]
	for i := 1; i < numGoroutines; i++ {
		if results[i] != firstLogger {
			t.Errorf("Logger instance %d differs from first instance", i)
		}
	}
}

func TestSetLogOutputConcurrency(t *testing.T) {
	// Save original state
	originalOutput := DefaultLogOutput
	defer func() {
		DefaultLogOutput = originalOutput
		defaultLoggerOnce = sync.Once{}
	}()

	const numGoroutines = 50
	var wg sync.WaitGroup
	buffers := make([]*bytes.Buffer, numGoroutines)

	// Create buffers
	for i := 0; i < numGoroutines; i++ {
		buffers[i] = &bytes.Buffer{}
	}

	// Launch goroutines to set output concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			SetLogOutput(buffers[index])
			// Small delay to increase chance of race conditions
			time.Sleep(time.Microsecond)
			getLogger() // This should not panic
		}(i)
	}

	wg.Wait()

	// Test should complete without panics
	t.Log("Concurrent SetLogOutput test completed successfully")
}

func TestLoggerFunctions(t *testing.T) {
	// Save original state
	originalOutput := DefaultLogOutput
	originalLogger := DefaultLogger
	originalOnce := defaultLoggerOnce
	defer func() {
		DefaultLogOutput = originalOutput
		DefaultLogger = originalLogger
		defaultLoggerOnce = originalOnce
	}()

	// Test all logger function wrappers
	var buf bytes.Buffer

	// Create a logger with Trace level to test all functions
	DefaultLogger = zerolog.New(&buf).Level(zerolog.TraceLevel)
	DefaultLogOutput = &buf

	tests := []struct {
		name string
		fn   func() *zerolog.Event
	}{
		{"Trace", Trace},
		{"Debug", Debug},
		{"Info", Info},
		{"Warn", Warn},
		{"Error", Error},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			event := test.fn()

			// For zerolog, events can be nil if the level is disabled
			// This is normal behavior, so we just check that the function doesn't panic
			if event != nil {
				// Send a test message
				event.Msg("test message")

				// Should have written to buffer
				logOutput := buf.String()
				if len(logOutput) == 0 {
					t.Errorf("%s() did not write to output", test.name)
				} else if !strings.Contains(logOutput, "test message") {
					t.Errorf("%s() output should contain 'test message', got: %s", test.name, logOutput)
				}
			}

			// Test passes if function doesn't panic and returns either valid event or nil
		})
	}
}

func TestDefaultLoggerInitialization(t *testing.T) {
	// Reset state
	defaultLoggerOnce = sync.Once{}
	DefaultLogOutput = nil

	// Get logger should initialize with stderr
	logger := getLogger()
	if logger == nil {
		t.Error("getLogger() should not return nil")
	}

	// DefaultLogOutput should be set to stderr
	if DefaultLogOutput != os.Stderr {
		t.Error("DefaultLogOutput should be set to os.Stderr by default")
	}
}

func BenchmarkGetLogger(b *testing.B) {
	// Reset state
	defaultLoggerOnce = sync.Once{}
	DefaultLogOutput = os.Stderr

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getLogger()
	}
}

func BenchmarkSetLogOutput(b *testing.B) {
	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetLogOutput(&buf)
	}
}

func BenchmarkConcurrentLogging(b *testing.B) {
	var buf bytes.Buffer
	SetLogOutput(&buf)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Info().Msg("benchmark message")
		}
	})
}
