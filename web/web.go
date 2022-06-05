package web

import (
	"log"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/NeilSeligmann/G15Manager/controller"
)

type WebServerInstance struct {
	WebServer       *gin.Engine
	SocketInstances map[uuid.UUID]*SocketInstance
	Dependencies    *controller.Dependencies
	IsLooping       bool
}

func NewWebServer(dep *controller.Dependencies) *WebServerInstance {
	webServerInstance := WebServerInstance{
		SocketInstances: make(map[uuid.UUID]*SocketInstance),
		Dependencies:    dep,
	}

	r := gin.Default()
	webServerInstance.WebServer = r

	var whitelist map[string]bool = map[string]bool{
		"127.0.0.1": true,
		"localhost": true,
	}

	r.Use(IPWhiteList(whitelist))

	// Serve web client
	r.Use(static.Serve("/", static.LocalFile("./data/web", true)))

	r.GET("/ping", func(c *gin.Context) {
		// dep.Volume.ToggleMuted()
		dep.Keyboard.BrightnessUp()

		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1 := r.Group("/v1")
	{
		// Web Socket
		v1.GET("/websocket", func(c *gin.Context) {
			uuid := uuid.New()
			instance := NewSocketInstance(&webServerInstance, uuid, c, dep)
			webServerInstance.SocketInstances[uuid] = &instance

			log.Printf("New socket connected: %s\n", uuid)

			defer instance.handleSocket(c)
			defer webServerInstance.StartLoop()
		})
	}

	go func() {
		r.Run(":34453") // listen and serve
	}()

	return &webServerInstance
}

func (webInst *WebServerInstance) StartLoop() {
	if webInst.IsLooping {
		return
	}

	go func() {
		for {
			// Stop loop
			if len(webInst.SocketInstances) < 1 {
				webInst.IsLooping = false
				break
			}

			// Wait 1 second
			time.Sleep(1 * time.Second)

			// Loop
			temps := webInst.Dependencies.Thermal.GetTemperatures()

			// Send temps to all sockets
			for _, socket := range webInst.SocketInstances {
				socket.SendTemperatures(temps)
			}
		}
	}()
}

func (webInst *WebServerInstance) onSocketClose(socketInst *SocketInstance) {
	// Removed instance from map
	delete(webInst.SocketInstances, socketInst.uuid)
}

func (webInst *WebServerInstance) BroadcastJSON(v interface{}) {
	for _, socketInst := range webInst.SocketInstances {
		socketInst.SendJSON(v)
	}
}

func (webInst *WebServerInstance) BroadcastInfo() {
	for _, socketInst := range webInst.SocketInstances {
		socketInst.ShouldSendInfo = true
	}
}
