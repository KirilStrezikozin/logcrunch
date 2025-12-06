// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/KirilStrezikozin/logcrunch/internal"
	"github.com/KirilStrezikozin/logcrunch/web/templates"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/rs/zerolog"
)

var addr = flag.String("source", "localhost:7779", "websocket source address")
var addrServe = flag.String("addr", "localhost:7780", "http service address")

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	url := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	c := internal.NewWebSocketClient(url, logger)
	s := internal.NewStore(1000)

	if err := c.Dial(); err != nil {
		logger.Fatal().Err(err).Msg("dial failed")
	}

	defer func() {
		<-c.Done()
		logger.Info().Msg("websocket client exited normally")
	}()

	go func() {
		err := c.Read(func(data []byte) {
			log, err := internal.NewLog(data)
			if err != nil {
				logger.Error().Err(err).Msg("invalid log data")
				return
			}
			s.AddLog(log)
			logger.Info().Str("id", string(log.ID)).Msg("log received")
		})
		if err != nil {
			logger.Error().Err(err).Msg("read")
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	fs := http.FileServer(http.Dir("web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.HandleFunc("/unread", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dataComponent := templates.Logs(s.GetUnreadLogs(1))
		err := dataComponent.Render(ctx, w)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})

	homeComponent := templates.Home("Logcrunch!")
	r.Handle("/", templ.Handler(homeComponent))

	server := http.Server{
		Addr:         *addrServe,
		Handler:      r,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	go func() {
		logger.Info().Msgf("starting HTTP server at %s", *addrServe)
		if err := server.ListenAndServe(); err != nil {
			logger.Error().Err(err).Msg("serve")
		}
	}()

	<-interrupt
	logger.Info().Msg("interrupt")

	if err := c.Close(); err != nil {
		logger.Error().Err(err).Msg("websocket client close")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("serve shutdown")
	} else {
		logger.Info().Msg("serve clean shutdown")
	}
}
