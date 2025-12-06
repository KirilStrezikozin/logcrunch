// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/KirilStrezikozin/logcrunch/internal"
	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
)

var upgrader = websocket.Upgrader{HandshakeTimeout: 10 * time.Second}

func ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	sendTicker := time.NewTicker(5 * time.Second)
	defer sendTicker.Stop()

	done := make(chan struct{})
	msgs := make(chan []byte, 100)

	s := internal.NewStore(10)
	s.AddLog(internal.Log{ID: "1", Message: "Hello, World!"})
	s.AddLog(internal.Log{ID: "2", Message: "Hello, World!"})
	s.AddLog(internal.Log{ID: "3", Message: "Hello, World!"})
	s.AddLog(internal.Log{ID: "4", Message: "Hello, World!"})
	s.AddLog(internal.Log{ID: "5", Message: "Hello, World!"})

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
			return
		case msg := <-msgs:
			err = c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-ticker.C:
			msgs <- []byte("ping")
		case <-sendTicker.C:
			logs := s.GetUnreadLogs(1)
			if len(logs) == 0 {
				continue
			}
			msg, err := logs[0].MarshalJSON()
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
