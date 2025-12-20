// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/KirilStrezikozin/logcrunch/internal"
	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
)

var upgrader = websocket.Upgrader{HandshakeTimeout: 10 * time.Second}

const (
	pongWait       = 10 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1 << 14
	writeWait      = 10 * time.Second
)

func ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	log.Print("new conn")

	c.SetReadLimit(maxMessageSize)
	c.SetReadDeadline(time.Now().Add(pongWait))
	c.SetPongHandler(func(_ string) error {
		log.Printf("pong received")
		c.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	sendTicker := time.NewTicker(5 * time.Second)
	pingTicker := time.NewTicker(pingPeriod)

	defer sendTicker.Stop()
	defer pingTicker.Stop()

	done := make(chan struct{})
	msgs := make(chan []byte, 100)
	lastLogID := -1

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)
		}
	}()

	for {
		select {
		case <-done:
			log.Println("closing")
			return
		case msg := <-msgs:
			c.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Println("write:", err)
				return
			}
		case <-pingTicker.C:
			c.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("ping:", err)
				return
			}
			log.Println("ping sent")
		case <-sendTicker.C:
			lastLogID++
			newLog := internal.Log{ID: internal.LogID(strconv.Itoa(lastLogID)), Message: "New log message"}
			msg, err := newLog.MarshalJSON()
			if err != nil {
				log.Println("marshal:", err)
				return
			}
			msgs <- msg
		}
	}
}

func main() {
	sourceHost := os.Getenv("LOGCRUNCH_SOURCE_HOST")
	sourcePort := os.Getenv("LOGCRUNCH_SOURCE_PORT")
	sourcePath := os.Getenv("LOGCRUNCH_SOURCE_PATH")

	http.HandleFunc(sourcePath, ws)
	log.Fatal(func() error {
		server := &http.Server{Addr: sourceHost + ":" + sourcePort, Handler: nil, WriteTimeout: 10 * time.Second, ReadTimeout: 10 * time.Second}
		return server.ListenAndServe()
	}())
}
