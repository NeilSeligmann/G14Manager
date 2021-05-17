package web

import (
	"github.com/gin-gonic/gin"

	"github.com/NeilSeligmann/G15Manager/controller"
	"github.com/NeilSeligmann/G15Manager/cxx/plugin/keyboard"
)

func NewHttpServer(dep *controller.Dependencies) *gin.Engine {
	r := gin.Default()

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
		v1.POST("/save", func(c *gin.Context) {
			dep.ConfigRegistry.Save()
		})

		// Web Socket
		v1.GET("/websocket", func(c *gin.Context) {
			instance := NewSocketInstance(c, dep)

			instance.handleSocket(c)
		})

		// Keyboard Routes
		kb := v1.Group("/keyboard")
		{
			kb.POST("/brightness", func(c *gin.Context) {
				increase := c.Query("increase")
				decrease := c.Query("decrease")
				value := c.Query("value")

				if increase == "true" {
					dep.Keyboard.BrightnessUp()
				} else if decrease == "true" {
					dep.Keyboard.BrightnessDown()
				} else if value != "" {
					var lvl = keyboard.OFF

					switch value {
					case "OFF":
						lvl = keyboard.OFF
					case "LOW":
						lvl = keyboard.LOW
					case "MEDIUM":
						lvl = keyboard.MEDIUM
					case "HIGH":
						lvl = keyboard.HIGH
					}

					dep.Keyboard.SetBrightness(lvl)
				} else {
					c.JSON(400, gin.H{
						"error": gin.H{
							"message": "No argument provided. Either 'increase', 'decrease', or 'value' are required",
						},
					})

					return
				}

				// Save
				dep.ConfigRegistry.Save()

				c.JSON(200, gin.H{
					"message": "succcess",
				})
			})

			// Enable/Disable touchpad
			kb.POST("/touchpad", func(c *gin.Context) {
				dep.Keyboard.ToggleTouchPad()

				c.JSON(200, gin.H{
					"message": "Successfully toggled touchpad",
				})
			})

			type RogKeyBody struct {
				Actions []string `json:"actions"`
			}

			// ROG key remap
			kb.POST("/rog", func(c *gin.Context) {
				var body RogKeyBody

				if c.ShouldBindJSON(&body) == nil {
					dep.Keyboard.Config.RogKey = body.Actions

					c.JSON(200, gin.H{
						"message": "Successfully set new ROG key bindings!",
					})
				} else {
					c.JSON(200, gin.H{
						"message": "Incorrect body provided!",
					})
				}
			})
		}

		// Thermal Routes
		thermal := v1.Group("/thermal")
		{
			profiles := thermal.Group("/profiles")
			{
				profiles.GET("/", func(c *gin.Context) {
					c.JSON(200, gin.H{
						"message":  "Successfully fetched all profiles!",
						"profiles": dep.Thermal.Profiles,
					})
				})

				profiles.GET("/current", func(c *gin.Context) {
					c.JSON(200, gin.H{
						"message":        "Profile successfully switched!",
						"currentProfile": dep.Thermal.CurrentProfile(),
					})
				})

				profiles.POST("/switch", func(c *gin.Context) {
					value := c.Query("value")
					dep.Thermal.SwitchToProfile(value)

					c.JSON(200, gin.H{
						"message": "Profile successfully switched!",
					})
				})
			}
		}
	}

	go func() {
		r.Run() // listen and serve on 0.0.0.0:8080
	}()

	return r
}

func NewWebSocketServer(dep *controller.Dependencies) {

}
