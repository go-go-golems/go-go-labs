package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type EventConfig struct {
	EventName string `json:"eventName"`
	Fields    struct {
		Location  bool `json:"location"`
		Timestamp bool `json:"timestamp"`
		Speed     bool `json:"speed"`
		Direction bool `json:"direction"`
		Size      bool `json:"size"`
	} `json:"fields"`
	Features struct {
		Validation bool `json:"validation"`
		Serializer bool `json:"serializer"`
		Kafka      bool `json:"kafka"`
	} `json:"features"`
}

func generateRubyCode(config EventConfig) string {
	var code bytes.Buffer
	code.WriteString("# frozen_string_literal: true\n\n")

	if config.Features.Kafka {
		code.WriteString("require 'kafka'\n")
	}

	code.WriteString(fmt.Sprintf("\nclass %sEvent\n", config.EventName))

	// Add attribute accessors
	var attrs []string
	if config.Fields.Location {
		attrs = append(attrs, "x", "y")
	}
	if config.Fields.Timestamp {
		attrs = append(attrs, "timestamp")
	}
	if config.Fields.Speed {
		attrs = append(attrs, "speed")
	}
	if config.Fields.Direction {
		attrs = append(attrs, "direction")
	}
	if config.Fields.Size {
		attrs = append(attrs, "size")
	}

	if len(attrs) > 0 {
		code.WriteString(fmt.Sprintf("  attr_accessor :%s\n\n", template.HTMLEscapeString(joinWithCommaColon(attrs))))
	}

	// Initialize
	code.WriteString(fmt.Sprintf("  def initialize(%s)\n", joinWithComma(attrs)))
	for _, attr := range attrs {
		code.WriteString(fmt.Sprintf("    @%s = %s\n", attr, attr))
	}
	code.WriteString("  end\n\n")

	// Add validation if requested
	if config.Features.Validation {
		generateValidation(&code, config, attrs)
	}

	// Add serializer if requested
	if config.Features.Serializer {
		generateSerializer(&code, attrs)
	}

	// Add Kafka producer if requested
	if config.Features.Kafka {
		generateKafkaProducer(&code)
	}

	code.WriteString("end\n")
	return code.String()
}

func joinWithComma(attrs []string) string {
	var result bytes.Buffer
	for i, attr := range attrs {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(attr)
	}
	return result.String()
}

func joinWithCommaColon(attrs []string) string {
	var result bytes.Buffer
	for i, attr := range attrs {
		if i > 0 {
			result.WriteString(", :")
		}
		result.WriteString(attr)
	}
	return result.String()
}

func generateValidation(code *bytes.Buffer, config EventConfig, attrs []string) {
	code.WriteString("  def valid?\n")
	code.WriteString(fmt.Sprintf("    return false if %s .nil?\n\n", joinWithNilCheck(attrs)))

	if config.Fields.Location {
		code.WriteString("    return false unless x.is_a?(Numeric) && y.is_a?(Numeric)\n")
	}
	if config.Fields.Speed {
		code.WriteString("    return false unless speed.is_a?(Numeric) && speed >= 0\n")
	}
	if config.Fields.Direction {
		code.WriteString("    return false unless (0..360).include?(direction)\n")
	}

	code.WriteString("    true\n")
	code.WriteString("  end\n\n")
}

func joinWithNilCheck(attrs []string) string {
	var result bytes.Buffer
	for i, attr := range attrs {
		if i > 0 {
			result.WriteString(" .nil? || ")
		}
		result.WriteString(attr)
	}
	return result.String()
}

func generateSerializer(code *bytes.Buffer, attrs []string) {
	code.WriteString("  def to_json\n")
	code.WriteString("    {\n")
	for _, attr := range attrs {
		code.WriteString(fmt.Sprintf("      %s: @%s,\n", attr, attr))
	}
	code.WriteString("    }.to_json\n")
	code.WriteString("  end\n\n")
}

func generateKafkaProducer(code *bytes.Buffer) {
	code.WriteString("  def publish\n")
	code.WriteString("    return unless valid?\n\n")
	code.WriteString("    kafka = Kafka.new(['localhost:9092'])\n")
	code.WriteString("    producer = kafka.producer\n")
	code.WriteString("    producer.produce(\n")
	code.WriteString("      topic: 'mouse_events',\n")
	code.WriteString("      payload: to_json\n")
	code.WriteString("    )\n")
	code.WriteString("    producer.deliver_messages\n")
	code.WriteString("  end\n")
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})

	http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var config EventConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		code := generateRubyCode(config)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(code))
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var config EventConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		code := generateRubyCode(config)
		w.Header().Set("Content-Type", "application/ruby")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_event.rb", config.EventName))
		w.Write([]byte(code))
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
