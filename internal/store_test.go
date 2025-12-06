// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package internal

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newLog(id int) Log {
	return Log{ID: LogID(strconv.Itoa(id))}
}

func newStore(initialCount int) *Store {
	s := NewStore(10)
	for i := range initialCount {
		s.AddLog(newLog(i))
	}
	return s
}

func TestStore_GetLogs(t *testing.T) {
	s := newStore(5)

	tests := []struct {
		name   string
		offset int
		limit  int
		expect []Log
	}{
		{"invalid offset", -1, 2, []Log{}},
		{"invalid limit", 1, 0, []Log{}},
		{"offset+limit >= len", 2, 4, []Log{}},
		{"valid window", 0, 2, []Log{newLog(3), newLog(4)}},
		{"valid window", 1, 2, []Log{newLog(2), newLog(3)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := s.GetLogs(tt.offset, tt.limit)
			assert.Equal(t, tt.expect, res)
		})
	}
}

func TestStore_GetUnreadLogs(t *testing.T) {
	t.Run("no logs", func(t *testing.T) {
		s := NewStore(10)
		res := s.GetUnreadLogs(3)
		assert.Len(t, res, 0)
	})

	t.Run("initial read all", func(t *testing.T) {
		s := newStore(4)
		assert.Equal(t, newStore(4).logs, s.GetUnreadLogs(10))
	})

	t.Run("no unread remaining", func(t *testing.T) {
		s := newStore(4)
		_ = s.GetUnreadLogs(10)
		assert.Len(t, s.GetUnreadLogs(5), 0)
	})

	t.Run("new logs partial then full", func(t *testing.T) {
		s := newStore(4)
		_ = s.GetUnreadLogs(10)

		s.AddLog(newLog(4))
		s.AddLog(newLog(5))

		assert.Equal(t, []Log{newLog(4)}, s.GetUnreadLogs(1))
		assert.Equal(t, []Log{newLog(5)}, s.GetUnreadLogs(10))
	})
}
