// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package internal

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type WebSocketError struct {
	Op  string
	Err error
}

func (e *WebSocketError) Error() string {
	return fmt.Sprintf("websocket client %s error: %v", e.Op, e.Err)
}

type WebSocketClient struct {
	url    url.URL
	conn   *websocket.Conn
	logger zerolog.Logger

	HandshakeTimeout time.Duration
	WriteTimeout     time.Duration
	ReadLimit        int64
}

func NewWebSocketClient(url url.URL, parentLogger zerolog.Logger) *WebSocketClient {
	logger := parentLogger.
		With().
		Str("server", url.String()).
		Logger()

	return &WebSocketClient{
		logger: logger,
		url:    url,

		HandshakeTimeout: 10 * time.Second,
		WriteTimeout:     10 * time.Second,
		ReadLimit:        1 << 14, // 16KB
	}
}

func (c *WebSocketClient) Dial() error {
	if c.conn != nil {
		return &WebSocketError{Op: "dial", Err: ErrConnectionAlreadyEstablished}
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: c.HandshakeTimeout,
	}

	var err error
	c.conn, _, err = dialer.Dial(c.url.String(), nil)
	if err != nil {
		return &WebSocketError{Op: "dial", Err: err}
	}

	c.conn.SetReadLimit(c.ReadLimit)
	c.conn.SetPingHandler(nil) // enable default ping handler

	c.logger.Info().Msg("connection established")
	return nil
}

func (c *WebSocketClient) Read(onRead func([]byte)) error {
	if c.conn == nil {
		return &WebSocketError{Op: "read", Err: ErrNilConnection}
	}

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return &WebSocketError{Op: "read", Err: err}
		}

		c.logger.Info().Msgf("recv: %s", msg)
		onRead(msg)
	}
}

func (c *WebSocketClient) Close() error {
	if c.conn == nil {
		return &WebSocketError{Op: "close", Err: ErrNilConnection}
	}

	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	c.conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	if err := c.conn.WriteMessage(websocket.CloseMessage, msg); err != nil {
		return &WebSocketError{Op: "close", Err: err}
	}

	if err := c.conn.Close(); err != nil {
		return &WebSocketError{Op: "close", Err: err}
	}

	c.logger.Info().Msg("connection closed")
	c.conn = nil
	return nil
}
