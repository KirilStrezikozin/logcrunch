// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package internal

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	WebSocketReadLimit        = 1 << 14 // 16KB
	WebSocketHandshakeTimeout = 10 * time.Second
	WebSocketWriteTimeout     = 10 * time.Second
)

type WebSocketError struct {
	Op  string
	Err error
}

func (e *WebSocketError) Error() string {
	return fmt.Sprintf("websocket client %s error: %v", e.Op, e.Err)
}

type IWebSocketControl interface {
	Dial(urlStr string) error
	Close() error
}

type IWebSocketReader interface {
	Read(onRead func(messageType int, p []byte)) error
}

type IWebSocketClient interface {
	IWebSocketControl
	IWebSocketReader
}

// XXX: WebSocketClient is not inherently thread-safe.
type WebSocketClient struct {
	conn *websocket.Conn

	logger zerolog.Logger

	HandshakeTimeout time.Duration
	WriteTimeout     time.Duration
	ReadLimit        int64
}

func NewWebSocketClient(string, parentLogger zerolog.Logger) *WebSocketClient {
	logger := parentLogger.
		With().
		Str("component", "websocket_client").
		Logger()

	return &WebSocketClient{
		logger: logger,

		HandshakeTimeout: WebSocketHandshakeTimeout,
		WriteTimeout:     WebSocketWriteTimeout,
		ReadLimit:        WebSocketReadLimit,
	}
}

func (c *WebSocketClient) Dial(urlStr string) error {
	if c.conn != nil {
		return &WebSocketError{Op: "dial", Err: ErrConnectionAlreadyEstablished}
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: c.HandshakeTimeout,
	}

	var err error
	c.conn, _, err = dialer.Dial(urlStr, nil)
	if err != nil {
		return &WebSocketError{Op: "dial", Err: err}
	}

	c.conn.SetReadLimit(c.ReadLimit)
	c.conn.SetPingHandler(nil) // enable default ping handler

	c.logger.Debug().Str("server", urlStr).Msg("connection established")
	return nil
}

func (c *WebSocketClient) Read(onRead func(messageType int, p []byte)) error {
	if c.conn == nil {
		return &WebSocketError{Op: "read", Err: ErrNilConnection}
	}

	urlStr := c.conn.RemoteAddr().String()

	for {
		messageType, p, err := c.conn.ReadMessage()
		if err != nil {
			return &WebSocketError{Op: "read", Err: err}
		}

		c.logger.Debug().Str("server", urlStr).Msg("message received")
		onRead(messageType, p)
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

	urlStr := c.conn.RemoteAddr().String()
	if err := c.conn.Close(); err != nil {
		return &WebSocketError{Op: "close", Err: err}
	}

	c.logger.Debug().Str("server", urlStr).Msg("connection closed")
	c.conn = nil
	return nil
}
