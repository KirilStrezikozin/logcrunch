// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:7779", "http service address")

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

	done := make(chan struct{})

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
		case <-ticker.C:
			err = c.WriteMessage(websocket.TextMessage, []byte("tick"))
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/ws", ws)
	log.Fatal(func() error {
		server := &http.Server{Addr: *addr, Handler: nil, WriteTimeout: 10 * time.Second, ReadTimeout: 10 * time.Second}
		return server.ListenAndServe()
	}())
}
