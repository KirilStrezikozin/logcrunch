// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package services

import (
	"github.com/KirilStrezikozin/logcrunch/internal"
	"github.com/rs/zerolog"
)

type ILogService interface {
	ReadLoop() error
}

type LogService struct {
	wsClient internal.IWebSocketReader
	logger   zerolog.Logger
}

func NewLogService(
	wsClient internal.IWebSocketReader,
	parentLogger zerolog.Logger,
) *LogService {
	logger := parentLogger.
		With().
		Str("service", "log").
		Logger()

	return &LogService{
		wsClient: wsClient,
		logger:   logger,
	}
}

func (s *LogService) ReadLoop() error {
	return s.wsClient.Read(func(messageType int, p []byte) {
		log, err := internal.NewLog(p)
		if err != nil {
			s.logger.Error().Err(err).Bytes("data", p).Msg("unparsable log data, skipping")
			return
		}

		s.logger.Debug().Str("id", string(log.ID)).Msg("log received")

		// Change reader service -> log service
		// s.store.AddLog(log)
	})
}
