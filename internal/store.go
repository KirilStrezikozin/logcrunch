// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package internal

import "sync"

// There is a number of logs that we store in memory, and the rest is stored in a db.
type Store struct {
	mu sync.RWMutex

	// we need a double buffer to read into one and save the other,
	// then truncate and flip to reuse memory periodically.
	// the third buffer keeps all logs for the current producer id.

	logs            []Log
	lastReadOffset  int
	lastSavedOffset int
}

func NewStore(capacity int) *Store {
	return &Store{
		logs:            make([]Log, 0, capacity),
		lastReadOffset:  -1,
		lastSavedOffset: -1,
	}
}

func (s *Store) AddLog(log Log) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, log)
}

func (s *Store) AddLogs(logs []Log) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, logs...)
}

func (s *Store) GetLogs(offset int, limit int) []Log {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if offset < 0 || limit <= 0 || offset+limit >= len(s.logs) {
		return s.logs[0:0] // empty
	}

	realOffset := len(s.logs) - offset - limit
	return s.logs[realOffset : realOffset+limit]
}

func (s *Store) GetUnreadLogs(limit int) []Log {
	s.mu.Lock()
	defer s.mu.Unlock()

	start := s.lastReadOffset + 1
	if start >= len(s.logs) || limit <= 0 {
		return s.logs[0:0] // empty
	}

	res := s.logs[start:min(start+limit, len(s.logs))]
	s.lastReadOffset += len(res)
	return res
}

func (s *Store) GetUnsavedLogs(limit int) []Log {
	s.mu.Lock()
	defer s.mu.Unlock()

	start := s.lastSavedOffset + 1
	if start >= len(s.logs) || limit <= 0 {
		return s.logs[0:0] // empty
	}

	res := s.logs[start:min(start+limit, len(s.logs))]
	s.lastSavedOffset += len(res)
	return res
}
