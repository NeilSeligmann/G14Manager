package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func IPWhiteList(whitelist map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !whitelist[c.ClientIP()] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"message": "Permission denied, address not in whitelist.",
				"data": gin.H{
					"clientIp": c.ClientIP(),
				},
			})
			return
		}
	}
}
