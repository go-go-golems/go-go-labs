package mp3lib

import (
	"testing"
	"time"
)

// MockWriter is a simple in-memory writer for testing purposes.
type MockWriter struct {
	data []byte
}

func (mw *MockWriter) Write(p []byte) (int, error) {
	mw.data = append(mw.data, p...)
	return len(p), nil
}

func TestSlowWriterWritesData(t *testing.T) {
	buffer := make([]byte, 4096) // 4 KB
	mw := &MockWriter{}
	sw := NewSlowWriter(mw, 1*time.Millisecond)

	n, err := sw.Write(buffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(buffer) {
		t.Fatalf("expected %d bytes written, got %d", len(buffer), n)
	}
	if len(mw.data) != len(buffer) {
		t.Fatalf("expected %d bytes in MockWriter, got %d", len(buffer), len(mw.data))
	}
}

func TestSlowWriterTakesExpectedTime(t *testing.T) {
	buffer := make([]byte, 4096) // 4 KB
	mw := &MockWriter{}
	sw := NewSlowWriter(mw, 10*time.Millisecond) // 10 ms per KB

	start := time.Now()
	_, err := sw.Write(buffer)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// We should sleep 4 times (since we have 4 KB), 10 ms each.
	expectedDuration := 40 * time.Millisecond
	if duration < expectedDuration {
		t.Fatalf("expected Write to take at least %v, but took %v", expectedDuration, duration)
	}
}

func TestSlowWriterHandlesSmallWrites(t *testing.T) {
	buffer := make([]byte, 500) // 500 B
	mw := &MockWriter{}
	sw := NewSlowWriter(mw, 10*time.Millisecond) // 10 ms per KB

	n, err := sw.Write(buffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(buffer) {
		t.Fatalf("expected %d bytes written, got %d", len(buffer), n)
	}
	if len(mw.data) != len(buffer) {
		t.Fatalf("expected %d bytes in MockWriter, got %d", len(buffer), len(mw.data))
	}
}
