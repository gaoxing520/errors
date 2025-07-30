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
	// Create a custom logger with buffer
	var buf bytes.Buffer
	customLogger := zerolog.New(&buf).Level(zerolog.InfoLevel)

	// Set the custom logger
	SetLogger(&customLogger)

	// Test that the logger was set by getting it
	logger := getLogger()
	if logger == nil {
		t.Error("Logger should not be nil")
	}

	// Test logging with custom logger
	logger.Info().Msg("test message")

	// The buffer should contain the log message
	logOutput := buf.String()
	if logOutput == "" {
		t.Error("Custom logger should have received log message")
	}

	// Verify the message contains our test content
	if !strings.Contains(logOutput, "test message") {
		t.Errorf("Log output should contain 'test message', got: %s", logOutput)
	}
}

func TestSetLogOutput(t *testing.T) {
	// Create a buffer for output
	var buf bytes.Buffer
	SetLogOutput(&buf)

	// Test that output was set
	if DefaultLogOutput != &buf {
		t.Error("DefaultLogOutput should be set to buffer")
	}

	// Get logger to initialize with new output
	logger := getLogger()

	// Test logging to custom output using Info level
	logger.Info().Msg("test output message")

	// The buffer should contain the log message
	logOutput := buf.String()
	if logOutput == "" {
		t.Error("Custom output should have received log message")
	}

	// Verify the message contains our test content
	if !strings.Contains(logOutput, "test output message") {
		t.Errorf("Log output should contain 'test output message', got: %s", logOutput)
	}
}

func TestGetLoggerConcurrency(t *testing.T) {
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

	// All results should be valid loggers
	for i := 0; i < numGoroutines; i++ {
		if results[i] == nil {
			t.Errorf("Logger instance %d is nil", i)
		}
	}
}

func TestSetLogOutputConcurrency(t *testing.T) {
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
	// Test all logger function wrappers
	var buf bytes.Buffer

	// Set trace level and output
	SetLogLevel(zerolog.TraceLevel)
	SetLogOutput(&buf)

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
				if logOutput == "" {
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
	// Save and restore original state
	originalOutput := DefaultLogOutput
	defer func() {
		SetLogOutput(originalOutput)
	}()

	// Reset to default
	SetLogOutput(os.Stderr)

	// Get logger should initialize with stderr
	logger := getLogger()
	if logger == nil {
		t.Error("getLogger() should not return nil")
	}

	// DefaultLogOutput should be set to stderr
	if DefaultLogOutput == os.Stderr {
		t.Log("DefaultLogOutput is correctly set to os.Stderr")
	} else {
		t.Error("DefaultLogOutput should be set to os.Stderr")
	}
}

func TestSetLogLevel(t *testing.T) {
	// Test with different log levels
	var buf bytes.Buffer
	SetLogOutput(&buf)

	// Test Info level (default)
	SetLogLevel(zerolog.InfoLevel)
	buf.Reset()
	Info().Msg("info message")
	if !strings.Contains(buf.String(), "info message") {
		t.Error("Info message should be logged at InfoLevel")
	}

	// Test Debug level (should not show at Info level)
	buf.Reset()
	Debug().Msg("debug message")
	if strings.Contains(buf.String(), "debug message") {
		t.Error("Debug message should not be logged at InfoLevel")
	}

	// Change to Debug level
	SetLogLevel(zerolog.DebugLevel)
	buf.Reset()
	Debug().Msg("debug message after level change")
	if !strings.Contains(buf.String(), "debug message after level change") {
		t.Error("Debug message should be logged after changing to DebugLevel")
	}

	// Test that Info still works
	buf.Reset()
	Info().Msg("info message after level change")
	if !strings.Contains(buf.String(), "info message after level change") {
		t.Error("Info message should still be logged at DebugLevel")
	}
}

func TestDefaultLogLevel(t *testing.T) {
	var buf bytes.Buffer
	SetLogOutput(&buf)
	SetLogLevel(zerolog.InfoLevel) // Reset to default

	// Test that Info level works by default
	buf.Reset()
	Info().Msg("default info test")
	if !strings.Contains(buf.String(), "default info test") {
		t.Error("Info should work with default InfoLevel")
	}

	// Test that Debug level doesn't work by default
	buf.Reset()
	Debug().Msg("default debug test")
	if strings.Contains(buf.String(), "default debug test") {
		t.Error("Debug should not work with default InfoLevel")
	}

	// Test that Error level works
	buf.Reset()
	Error().Msg("default error test")
	if !strings.Contains(buf.String(), "default error test") {
		t.Error("Error should work with default InfoLevel")
	}
}

func BenchmarkGetLogger(b *testing.B) {
	// Ensure logger is initialized first
	SetLogOutput(os.Stderr)

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
	// Use os.Stderr for concurrent logging to avoid buffer race conditions
	SetLogOutput(os.Stderr)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Info().Msg("benchmark message")
		}
	})
}
