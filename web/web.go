package web

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/NeilSeligmann/G15Manager/controller"
)

type WebServerInstance struct {
	WebServer       *gin.Engine
	SocketInstances map[uuid.UUID]*SocketInstance
}

func NewWebServer(dep *controller.Dependencies) *WebServerInstance {
	webServerInstance := WebServerInstance{
		SocketInstances: make(map[uuid.UUID]*SocketInstance),
	}

	r := gin.Default()
	webServerInstance.WebServer = r

	var whitelist map[string]bool = map[string]bool{
		"127.0.0.1": true,
	}

	r.Use(IPWhiteList(whitelist))

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

			instance.handleSocket(c)
		})
	}

	go func() {
		r.Run() // listen and serve on 0.0.0.0:8080
	}()

	return &webServerInstance
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
		// socketInst.SendInfo()
		socketInst.ShouldSendInfo = true
	}
}
