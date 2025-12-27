// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLog_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": 1,
		"timestamp": 123456,
		"level": "info",
		"message": "hello world",
		"source_file": "main",
		"source_line": 11,
		"source_function": "Class.TestFunc",
		"function_call_started_at": 100.0,
		"function_call_ended_at": 200.0,
		"function_duration": 100.0,
		"call_stack": [1234],
		"attrs": {
			"user": "alice",
			"count": 42,
			"nested": {
				"flag": true
			},
			"list": [1,2,3]
		}
	}`

	log, err := NewLog([]byte(jsonData))
	assert.NoError(t, err)

	assert.Equal(t, LogID(1), log.ID)
	assert.Equal(t, Timestamp(123456), log.Timestamp)
	assert.Equal(t, "info", log.Level)
	assert.Equal(t, "hello world", log.Message)
	assert.Equal(t, "main", log.SourceFile)
	assert.Equal(t, 11, log.SourceLine)
	assert.Equal(t, "Class.TestFunc", log.SourceFunction)
	assert.Equal(t, Timestamp(100), log.FunctionCallStartedAt)
	assert.Equal(t, Timestamp(200), log.FunctionCallEndedAt)
	assert.Equal(t, []LogID{1234}, log.FunctionCallStack)

	assert.Equal(t, "alice", log.Attrs["user"])
	assert.Equal(t, float64(42), log.Attrs["count"])
	nested, ok := log.Attrs["nested"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, true, nested["flag"])

	assert.Equal(t, LogID(1), log.parsedAttrs["id"])
	assert.Equal(t, "alice", log.parsedAttrs["attrs.user"])
	assert.Equal(t, float64(42), log.parsedAttrs["attrs.count"])
	assert.Equal(t, true, log.parsedAttrs["attrs.nested.flag"])
	assert.Nil(t, log.parsedAttrs["attrs.list"])
}

func TestLog_Type(t *testing.T) {
	infoLog := Log{
		ID:        1,
		Level:     "info",
		Message:   "info message",
		Timestamp: 123,
	}
	assert.Equal(t, LogTypeInfo, infoLog.Type())

	metricLog := Log{
		ID:                    2,
		Level:                 "info",
		Message:               "metric message",
		Timestamp:             123,
		FunctionCallStartedAt: 1,
		FunctionCallEndedAt:   2,
	}
	assert.Equal(t, LogTypeMetric, metricLog.Type())
}

func TestParseAttrsRecursive_Empty(t *testing.T) {
	var log Log
	log.parseAttrs()
	assert.Equal(t, LogID(0), log.parsedAttrs["id"])
	assert.Equal(t, Timestamp(0), log.parsedAttrs["timestamp"])
	assert.Nil(t, log.parsedAttrs["call_stack"])
}

func TestParseAttrsRecursive_Nested(t *testing.T) {
	log := Log{
		Attrs: map[string]any{
			"a": map[string]any{
				"b": float64(42),
			},
			"d": float64(1),
			"c": "str",
		},
	}
	log.parseAttrs()
	assert.Equal(t, float64(42), log.parsedAttrs["attrs.a.b"])
	assert.Equal(t, float64(1), log.parsedAttrs["attrs.d"])
	assert.Equal(t, "str", log.parsedAttrs["attrs.c"])
}
