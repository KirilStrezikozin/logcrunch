// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package internal

// TODO: there could be a number of logs that we store in memory, and the rest is stored in a db.
// TODO: use RWMutext

type Store struct {
	logs           []Log
	lastReadOffset int
}

func NewStore(capacity int) *Store {
	return &Store{
		logs:           make([]Log, 0, capacity),
		lastReadOffset: -1,
	}
}

func (s *Store) AddLog(log Log) {
	s.logs = append(s.logs, log)
}

func (s *Store) GetLogs(offset int, limit int) []Log {
	if offset < 0 || limit <= 0 || offset+limit >= len(s.logs) {
		return s.logs[0:0] // empty
	}

	realOffset := len(s.logs) - offset - limit
	return s.logs[realOffset : realOffset+limit]
}

func (s *Store) GetUnreadLogs(limit int) []Log {
	start := s.lastReadOffset + 1
	if start >= len(s.logs) || limit <= 0 {
		return s.logs[0:0] // empty
	}

	res := s.logs[start:min(start+limit, len(s.logs))]
	s.lastReadOffset += len(res)
	return res
}
