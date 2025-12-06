// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package internal

import (
	"encoding/json"
	"fmt"
)

type LogID string
type Timestamp float64
type LogType int

const (
	LogTypeInfo LogType = iota + 1
	LogTypeMetric
)

type Log struct {
	ID        LogID     `json:"id"`
	Timestamp Timestamp `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`

	// Where the log was generated.
	Namespace    string `json:"namespace,omitempty"`
	ClassName    string `json:"class_name,omitempty"`
	FunctionName string `json:"function_name,omitempty"`

	// Only relevant for logs that describe function calls
	// and include performance data.
	FunctionCallStartedAt Timestamp `json:"function_call_started_at,omitempty"`
	FunctionCallEndedAt   Timestamp `json:"function_call_ended_at,omitempty"`
	FunctionDuration      float64   `json:"function_duration,omitempty"`

	// Call stack leading to the log describing a function call.
	FunctionCallStack []LogID `json:"call_stack,omitempty"`

	// Context attributes associated with the log.
	Attrs map[string]any `json:"attrs,omitempty"`

	// Map of log attribute path to value for quick access.
	parsedAttrs map[string]any
}

func NewLog(data []byte) (Log, error) {
	var log Log
	err := json.Unmarshal(data, &log)
	if err != nil {
		return log, fmt.Errorf("error unmarshaling log data: %w", err)
	}

	log.parseAttrs()

	return log, nil
}

func (l *Log) Type() LogType {
	if l.FunctionCallStartedAt != 0 && l.FunctionCallEndedAt != 0 {
		return LogTypeMetric
	}
	return LogTypeInfo
}

func (l *Log) parseAttrs() {
	parsed := make(map[string]any)

	parsed["id"] = l.ID
	parsed["timestamp"] = l.Timestamp
	parsed["level"] = l.Level
	parsed["message"] = l.Message
	parsed["namespace"] = l.Namespace
	parsed["class_name"] = l.ClassName
	parsed["function_name"] = l.FunctionName
	parsed["function_call_started_at"] = l.FunctionCallStartedAt
	parsed["function_call_ended_at"] = l.FunctionCallEndedAt
	parsed["function_duration"] = l.FunctionDuration
	parsed["call_stack"] = nil

	parseAttrsRecursive(l.Attrs, parsed, "attrs")

	l.parsedAttrs = parsed
}

func parseAttrsRecursive(attrs map[string]any, dest map[string]any, prefix string) {
	if attrs == nil {
		return
	}

	for k, v := range attrs {
		var path string
		if prefix != "" {
			path = prefix + "." + k
		} else {
			path = k
		}

		switch val := v.(type) {
		case string, float64, bool:
			dest[path] = val
		case []any:
			dest[path] = nil
		case map[string]any:
			parseAttrsRecursive(val, dest, path)
		}
	}
}

func (l *Log) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(*l)
	if err != nil {
		return nil, fmt.Errorf("error marshaling log data: %w", err)
	}
	return data, nil
}
