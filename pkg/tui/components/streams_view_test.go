package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/pkg/tui/models"
	"github.com/go-go-golems/go-go-labs/pkg/tui/styles"
)

func TestStreamsView_Init(t *testing.T) {
	view := NewStreamsView(styles.NewStyles())
	cmd := view.Init()
	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

func TestStreamsView_UpdateWithStreamsData(t *testing.T) {
	view := NewStreamsView(styles.NewStyles())

	streams := []models.StreamData{
		{
			Name:         "test-stream",
			Length:       1000,
			MemoryUsage:  1024,
			Groups:       2,
			LastID:       "123-0",
			MessageRates: []float64{0.5, 0.7, 0.9},
		},
	}

	msg := StreamsDataMsg{Streams: streams}
	updatedView, cmd := view.Update(msg)

	if cmd != nil {
		t.Error("Update should return nil command for data message")
	}

	streamsView := updatedView.(*StreamsView)
	if len(streamsView.streams) != 1 {
		t.Errorf("Expected 1 stream, got %d", len(streamsView.streams))
	}

	if streamsView.streams[0].Name != "test-stream" {
		t.Errorf("Expected stream name 'test-stream', got %s", streamsView.streams[0].Name)
	}
}

func TestStreamsView_KeyNavigation(t *testing.T) {
	view := NewStreamsView(styles.NewStyles())

	// Add some test data
	streams := []models.StreamData{
		{Name: "stream1"},
		{Name: "stream2"},
		{Name: "stream3"},
	}
	view.Update(StreamsDataMsg{Streams: streams})

	// Test down key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedView, _ := view.Update(keyMsg)
	streamsView := updatedView.(*StreamsView)

	if streamsView.selectedIdx != 1 {
		t.Errorf("Expected selectedIdx to be 1, got %d", streamsView.selectedIdx)
	}

	// Test up key
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updatedView, _ = streamsView.Update(keyMsg)
	streamsView = updatedView.(*StreamsView)

	if streamsView.selectedIdx != 0 {
		t.Errorf("Expected selectedIdx to be 0, got %d", streamsView.selectedIdx)
	}
}

func TestStreamsView_ViewRendering(t *testing.T) {
	view := NewStreamsView(styles.NewStyles())

	// Test empty view
	output := view.View()
	if output == "" {
		t.Error("View should return non-empty string even when no data")
	}

	// Add test data
	streams := []models.StreamData{
		{
			Name:         "test-stream",
			Length:       1000,
			MemoryUsage:  1024,
			Groups:       2,
			LastID:       "123-0",
			MessageRates: []float64{0.5, 0.7, 0.9},
		},
	}
	view.Update(StreamsDataMsg{Streams: streams})

	output = view.View()
	if output == "" {
		t.Error("View should return non-empty string with data")
	}
}

func TestStreamsView_GetSelectedStream(t *testing.T) {
	view := NewStreamsView(styles.NewStyles())

	// Test with no data
	selected := view.GetSelectedStream()
	if selected != nil {
		t.Error("GetSelectedStream should return nil when no data")
	}

	// Add test data
	streams := []models.StreamData{
		{Name: "stream1"},
		{Name: "stream2"},
	}
	view.Update(StreamsDataMsg{Streams: streams})

	selected = view.GetSelectedStream()
	if selected == nil {
		t.Error("GetSelectedStream should return stream when data exists")
	}

	if selected.Name != "stream1" {
		t.Errorf("Expected selected stream name 'stream1', got %s", selected.Name)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, test := range tests {
		result := formatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", test.bytes, result, test.expected)
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"short", 10, "short"},
		{"exact length", 12, "exact length"},
		{"this is a very long string", 10, "this is..."},
		{"medium", 5, "me..."},
	}

	for _, test := range tests {
		result := truncateString(test.input, test.length)
		if result != test.expected {
			t.Errorf("truncateString(%s, %d) = %s, expected %s", test.input, test.length, result, test.expected)
		}
	}
}
