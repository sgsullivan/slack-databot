package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
)

type websocketData struct {
	ws *websocket.Conn
}

// connectWebsocket ...
func connectWebsocket(wssUrl string) *websocket.Conn {
	ws, err := websocket.Dial(wssUrl, "", "http://localhost/")
	if err != nil {
		log.Fatal(fmt.Sprintf("Error connecting to websocket: %s", err))
	}
	return ws
}

// readSocket ...
func (wsClient *websocketData) readSocket() []byte {
	msg := make([]byte, slackMsgSizeCapBytes)
	n, err := wsClient.ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}
	logDebug(fmt.Sprintf("Received %d bytes", n))
	return msg
}

// writeSocket ...
func (wsClient *websocketData) writeSocket(data []byte) {
	n, err := wsClient.ws.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	logDebug(fmt.Sprintf("Wrote %d bytes", n))
}
