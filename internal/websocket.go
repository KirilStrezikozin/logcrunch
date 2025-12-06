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
	done   chan struct{}

	HandshakeTimeout time.Duration
	WriteTimeout     time.Duration
}

func NewWebSocketClient(url url.URL, parentLogger zerolog.Logger) *WebSocketClient {
	logger := parentLogger.
		With().
		Str("server", url.String()).
		Logger()

	return &WebSocketClient{
		logger: logger,
		url:    url,
		done:   make(chan struct{}),

		HandshakeTimeout: 10 * time.Second,
		WriteTimeout:     10 * time.Second,
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

	c.logger.Info().Msg("connection established")
	return nil
}

func (c *WebSocketClient) Read(onRead func([]byte)) error {
	if c.conn == nil {
		return &WebSocketError{Op: "read", Err: ErrNilConnection}
	}

	defer close(c.done)
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

	<-c.done
	if err := c.conn.Close(); err != nil {
		return &WebSocketError{Op: "close", Err: err}
	}

	c.logger.Info().Msg("connection closed")
	return nil
}

func (c *WebSocketClient) Done() <-chan struct{} {
	return c.done
}
