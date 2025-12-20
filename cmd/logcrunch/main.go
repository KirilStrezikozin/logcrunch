// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/KirilStrezikozin/logcrunch/internal"
	"github.com/KirilStrezikozin/logcrunch/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/joho/godotenv/autoload"

	"github.com/rs/zerolog"
)

func main() {
	sourceScheme := os.Getenv("LOGCRUNCH_SOURCE_SCHEME")
	sourceHost := os.Getenv("LOGCRUNCH_SOURCE_HOST")
	sourcePort := os.Getenv("LOGCRUNCH_SOURCE_PORT")
	sourcePath := os.Getenv("LOGCRUNCH_SOURCE_PATH")
	serveHost := os.Getenv("LOGCRUNCH_SERVE_HOST")
	servePort := os.Getenv("LOGCRUNCH_SERVE_PORT")

	reconnect := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	url := url.URL{Scheme: sourceScheme, Host: sourceHost + ":" + sourcePort, Path: sourcePath}
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Logger()

	c := internal.NewWebSocketClient(url, logger)
	s := internal.NewStore(1000)

	readLoop := func() error {
		return c.Read(func(data []byte) {
			log, err := internal.NewLog(data)
			if err != nil {
				logger.Error().Err(err).Msg("invalid log data")
				return
			}

			logger.Info().Str("id", string(log.ID)).Msg("log received")
			s.AddLog(log)
		})
	}

	doReconnect := func() {
		defer func() { reconnect <- struct{}{} }()

		if err := c.Dial(); err != nil {
			logger.Error().Err(err).Msg("dial failed")
			return
		}

		if err := readLoop(); err != nil {
			logger.Error().Err(err).Msg("read")
		}
	}

	reqLogger := middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: &logger,
	})

	h := internal.NewHandler(logger)

	r := chi.NewRouter()
	r.Use(reqLogger)

	r.Handle(types.EndpointStatic, h.Static())
	r.Handle(types.EndpontIndex, h.Index())

	r.Get(types.EndpointGetConnectionStatus, h.GetConnectionStatus)
	r.Get(types.EndpointGetConnectionURL, h.GetConnectionURL)
	r.Post(types.EndpointPostConnectionURL, h.PostConnectionURL)

	r.HandleFunc("/unread", func(w http.ResponseWriter, r *http.Request) {
		// ctx := r.Context()
		// dataComponent := templates.Logs(s.GetUnreadLogs(1))
		// err := dataComponent.Render(ctx, w)
		// if err != nil {
		// 	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 	return
		// }
	})

	server := http.Server{
		Addr:         serveHost + ":" + servePort,
		Handler:      r,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	go func() {
		logger.Info().Msgf("starting HTTP server at %s:%s", serveHost, servePort)
		if err := server.ListenAndServe(); err != nil {
			logger.Error().Err(err).Msg("serve")
		}
	}()

	go doReconnect() // initial connect

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error().Err(err).Msg("serve shutdown")
		} else {
			logger.Info().Msg("serve clean shutdown")
		}
	}()

	for {
		select {
		case <-interrupt:
			logger.Info().Msg("interrupt")
			if err := c.Close(); err != nil {
				logger.Error().Err(err).Msg("websocket client close")
			}
			<-reconnect // wait for the reconnect go-routine to exit
			return
		case <-reconnect:
			if err := c.Close(); err != nil {
				logger.Error().Err(err).Msg("websocket client close")
			}
			logger.Info().Msgf("(re)connecting to %s in 3 seconds...", url.String())
			select {
			case <-interrupt:
				logger.Info().Msg("reconnect cancelled due to interrupt")
				return
			case <-time.After(3 * time.Second):
				go doReconnect()
			}
		}
	}
}
