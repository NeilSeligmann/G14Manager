package web

import (
	"github.com/gin-gonic/gin"

	"github.com/zllovesuki/G14Manager/controller"
)

// type HttpServerStruct struct {

// }

// type Notifier struct {
// 	C    chan util.Notification
// 	show chan string
// 	hide chan struct{}
// }

func NewHttpServer(dep *controller.Dependencies) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		// dep.Volume.ToggleMuted()
		dep.Keyboard.BrightnessUp()

		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// r.Run() // listen and serve on 0.0.0.0:8080

	go func() {
		print("TEST 1!")
		r.Run()
		print("TEST 2!")
	}()
	print("TEST 123!")

	return r
}
