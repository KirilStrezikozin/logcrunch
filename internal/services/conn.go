// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package services

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/KirilStrezikozin/logcrunch/internal"
	"github.com/KirilStrezikozin/logcrunch/internal/types"
	"github.com/KirilStrezikozin/logcrunch/pkg/strings"
	"github.com/rs/zerolog"
)

type IConnectionService interface {
	GetURL() (string, error)
	SetURL(url string) (string, error)
	GetStatus() (types.ConnectionStatus, error)

	ConnectOnce()
	ReconnectLoop(interrupt <-chan struct{})
}

type ConnectionService struct {
	mu   sync.Mutex
	once sync.Once

	doConnect   chan struct{}
	connectDone chan struct{}

	url    strings.Buffer
	status atomic.Int32

	db         internal.DBReadWriter
	wsClient   internal.IWebSocketControl
	logService ILogService
	logger     zerolog.Logger
}

func NewConnectionService(
	db internal.DBReadWriter,
	wsClient internal.IWebSocketControl,
	logService ILogService,
	parentLogger zerolog.Logger,
) *ConnectionService {
	logger := parentLogger.
		With().
		Str("service", "connection").
		Logger()

	s := &ConnectionService{
		doConnect:   make(chan struct{}, 1),
		connectDone: make(chan struct{}),

		db:         db,
		wsClient:   wsClient,
		logService: logService,
		logger:     logger,
	}

	s.status.Store(int32(types.ConnectionStatusDisconnected))
	return s
}

// XXX: unsafely modifying the returned string will modify the internal buffer.
func (s *ConnectionService) GetURL() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.db.Get(types.GetConnectionBucketName(), types.GetConnectionURLKey(), func(value []byte) error {
		if value == nil {
			return nil
		}

		// Avoid string allocations by reusing the internal buffer.
		s.url.Reset()
		_, err := s.url.Write(value)
		if err != nil {
			panic(err) // Writing to a strings.Buffer should never fail.
		}

		return nil
	})

	val := s.url.String()
	if err != nil {
		return val, fmt.Errorf("failed to get connection url from db: %w", err)
	}
	return val, nil
}

func (s *ConnectionService) SetURL(url string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.db.Put(types.GetConnectionBucketName(), types.GetConnectionURLKey(), []byte(url))
	if err != nil {
		return s.url.String(), fmt.Errorf("failed to put connection url to db: %w", err)
	}

	s.url.Reset()
	_, err = s.url.WriteString(url)
	if err != nil {
		panic(err) // Writing to a strings.Buffer should never fail.
	}

	s.triggerConnect()
	return s.url.String(), nil
}

func (s *ConnectionService) GetStatus() (types.ConnectionStatus, error) {
	return types.ConnectionStatus(s.status.Load()), nil
}

func (s *ConnectionService) ConnectOnce() {
	s.once.Do(s.triggerConnect)
}

func (s *ConnectionService) triggerConnect() {
	s.status.Store(int32(types.ConnectionStatusConnecting))
	s.doConnect <- struct{}{}
}

func (s *ConnectionService) connect() {
	defer func() { s.connectDone <- struct{}{} }()
	s.status.Store(int32(types.ConnectionStatusConnecting))

	urlStr, err := s.GetURL()
	if err != nil {
		s.status.Store(int32(types.ConnectionStatusError))
		s.logger.Error().Err(err).Msg("failed to get connection url")
		return
	}

	if err := s.wsClient.Dial(urlStr); err != nil {
		s.status.Store(int32(types.ConnectionStatusError))
		s.logger.Error().Err(err).Msg("dial failed")
		return
	}

	s.status.Store(int32(types.ConnectionStatusConnected))

	if err := s.logService.ReadLoop(); err != nil {
		s.status.Store(int32(types.ConnectionStatusDisconnected))
		s.logger.Error().Err(err).Msg("read")
		return
	}

	s.status.Store(int32(types.ConnectionStatusDisconnected))
}

const ReconnectDelay = 3 * time.Second

func (s *ConnectionService) ReconnectLoop(interrupt <-chan struct{}) {
	cancel := func() {
		// Close existing connection if any.
		if err := s.wsClient.Close(); err != nil {
			s.logger.Debug().Err(err).Msg("websocket client close failed")
		}

		<-s.connectDone // Wait for the connect go-routine to exit.
		s.status.Store(int32(types.ConnectionStatusDisconnected))
	}

	for {
		select {
		case <-interrupt:
			s.logger.Debug().Msg("reconnect loop interrupted, stopping...")
			cancel()
			return
		case <-s.doConnect:
			s.logger.Info().Msgf("connecting to %s...", s.url.String())
			cancel()
			go s.connect()
		case <-s.connectDone:
			if err := s.wsClient.Close(); err != nil {
				s.logger.Debug().Err(err).Msg("websocket client close failed")
			}

			s.status.Store(int32(types.ConnectionStatusConnecting))
			s.logger.Info().Msgf("reconnecting to %s in 3 seconds...", s.url.String())

			select {
			case <-interrupt:
				s.status.Store(int32(types.ConnectionStatusDisconnected))
				s.logger.Debug().Msg("reconnect loop interrupted, stopping...")
				return
			case <-time.After(ReconnectDelay):
				go s.connect()
			}
		}
	}
}
