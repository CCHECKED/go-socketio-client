package main

import (
	"crypto/tls"
	"fmt"
	"github.com/CCHECKED/socketio-client"
	"github.com/CCHECKED/socketio-client/consts"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	manager := socketio.NewSocketManager(
		&url.URL{
			Scheme: "https",
			Host:   "localhost:3000",
			Path:   "/devices/",
		},
		&socketio.SocketManagerConfig{
			AllowUpgrade: false,
			Tls:          &tls.Config{},
			Headers:      &http.Header{},
			Debug:        false,
		},
	)

	// Подключение к namespace "/"
	client := manager.DefaultSocket("/")
	// Подключение к namespace "/chat"
	clientChat := manager.DefaultSocket("/chat")

	client.On("message", func(data interface{}) {
		fmt.Println("Recieved message:", data)
	})

	client.OnService(consts.PING, func() {
		fmt.Println("Recieved Ping")
	})

	clientChat.On("message", func(data interface{}) {
		fmt.Println("Recieved chat message:", data)
	})

	termSign := make(chan os.Signal, 1)
	signal.Notify(termSign, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-termSign:
			return
		}
	}
}
