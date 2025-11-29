// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"net/url"
	"os"
	"os/signal"

	"github.com/KirilStrezikozin/logcrunch/internal"
	"github.com/rs/zerolog"
)

var addr = flag.String("addr", "localhost:7779", "http service address")

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	url := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	c := internal.NewWebSocketClient(url, logger)

	if err := c.Dial(); err != nil {
		logger.Fatal().Err(err).Msg("dial failed")
	}

	defer c.Close()

	go func() {
		c.Read(func(b []byte) {
			// Nothing.
		})
	}()

	for {
		select {
		case <-c.Done():
			return
		case <-interrupt:
			logger.Info().Msg("interrupt")
			return
		}
	}
}
