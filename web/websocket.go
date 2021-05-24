package web

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/NeilSeligmann/G15Manager/controller"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// func ping(ws *websocket.Conn, done chan struct{}) {
// 	ticker := time.NewTicker(pingPeriod)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case <-ticker.C:
// 			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
// 				log.Println("ping:", err)
// 			}
// 		case <-done:
// 			return
// 		}
// 	}
// }

type SocketInstance struct {
	WebServer      *WebServerInstance
	uuid           uuid.UUID
	Context        *gin.Context
	Dependencies   *controller.Dependencies
	ShouldSendInfo bool
	ws             *websocket.Conn
}

func NewSocketInstance(webServer *WebServerInstance, uuid uuid.UUID, c *gin.Context, dep *controller.Dependencies) SocketInstance {
	instance := SocketInstance{
		WebServer:      webServer,
		uuid:           uuid,
		Context:        c,
		Dependencies:   dep,
		ShouldSendInfo: true,
	}

	return instance
}

type SocketMessage struct {
	Category int    `json:"category"`
	Action   int    `json:"action"`
	Value    string `json:"value"`
}

func (inst *SocketInstance) handleSocket(c *gin.Context) {
	//Upgrade get request to webSocket protocol
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("Error during connection upgrade:", err)
		return
	}
	defer ws.Close()

	inst.ws = ws

	// The event loop
	for {
		if inst.ShouldSendInfo {
			inst.ShouldSendInfo = false
			inst.SendInfo()
		}

		messageType, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error during message reading:", err)
			ws.Close()
			inst.WebServer.onSocketClose(inst)
			break
		}

		inst.processMessage(messageType, message)
	}
}

func (inst *SocketInstance) processMessage(messageType int, message []byte) {
	if string(message) == "hb" {
		err := inst.ws.WriteMessage(messageType, []byte("alive"))
		if err != nil {
			log.Println("Error responding to heartbeat:", err)
		}

		return
	}

	// log.Printf("Received: %s", message)

	decodedMessage := SocketMessage{}

	err := json.Unmarshal(message, &decodedMessage)
	if err != nil {
		if err != nil {
			log.Println("Failed to parse JSON:", err)
			return
		}
	}

	// log.Printf("decodedMessage")
	// log.Print(decodedMessage)

	switch decodedMessage.Category {
	// Info
	case 0:
		inst.handleSystemMessage(decodedMessage.Action, decodedMessage.Value)
	// Thermal
	case 1:
		inst.Dependencies.Thermal.HandleWSMessage(inst.ws, decodedMessage.Action, decodedMessage.Value)
	// Keyboard
	case 2:
		inst.Dependencies.Keyboard.HandleWSMessage(inst.ws, decodedMessage.Action, decodedMessage.Value)
	// Battery
	case 3:
		inst.Dependencies.Battery.HandleWSMessage(inst.ws, decodedMessage.Action, decodedMessage.Value)
	// RR
	case 4:
		inst.Dependencies.RR.HandleWSMessage(inst.ws, decodedMessage.Action, decodedMessage.Value)
	// Volume
	case 5:
		inst.Dependencies.Volume.HandleWSMessage(inst.ws, decodedMessage.Action, decodedMessage.Value)
	}

	// Save config
	inst.Dependencies.ConfigRegistry.Save()

	// Send update info
	inst.SendInfo()
}

func (inst *SocketInstance) handleSystemMessage(action int, value string) {
	switch action {
	// Info
	case 0:
		inst.SendInfo()
	}
}

func (inst *SocketInstance) SendInfo() {
	inst.SendJSON(gin.H{
		"action": 0,
		"data": gin.H{
			"thermal":  inst.Dependencies.Thermal.GetWSInfo(),
			"keyboard": inst.Dependencies.Keyboard.GetWSInfo(),
			"volume":   inst.Dependencies.Volume.GetWSInfo(),
			"rr":       inst.Dependencies.RR.GetWSInfo(),
			"battery":  inst.Dependencies.Battery.GetWSInfo(),
		},
	})
}

func (inst *SocketInstance) SendJSON(v interface{}) {
	err := inst.ws.WriteJSON(v)

	if err != nil {
		log.Printf("Failed to send JSON message!")
		log.Print(v)
	}
}
