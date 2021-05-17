package web

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Time to wait before force close on connection.
	closeGracePeriod = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

// test 123

func ping(ws *websocket.Conn, done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				log.Println("ping:", err)
			}
		case <-done:
			return
		}
	}
}

type SocketMessage struct {
	Action int    `json:"action"`
	Value  string `json:"value"`
}

func socketHandler(c *gin.Context) {
	//Upgrade get request to webSocket protocol
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("Error during connection upgrade:", err)
		return
	}
	defer ws.Close()

	// The event loop
	for {
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error during message reading:", err)
			break
		}

		// go func() {
		processMessage(ws, messageType, message)
		// }()
	}
}

func processMessage(ws *websocket.Conn, messageType int, message []byte) {
	if string(message) == "heartbeat" {
		// err = ws.WriteMessage(websocket.TextMessage, []byte("alive"))
		err := ws.WriteMessage(messageType, []byte("alive"))
		if err != nil {
			log.Println("Error responding to heartbeat:", err)
		}

		return
	}

	log.Printf("Received: %s", message)

	decodedMessage := SocketMessage{}

	err := json.Unmarshal(message, &decodedMessage)
	if err != nil {
		if err != nil {
			log.Println("Failed to parse JSON:", err)
			return
		}
	}

	log.Printf("decodedMessage")
	log.Print(decodedMessage)

	// err = ws.WriteMessage(messageType, message)
	// if err != nil {
	// 	log.Println("Error during message writing:", err)
	// 	break
	// }
}
