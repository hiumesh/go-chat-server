package api

import (
	"github.com/gin-gonic/gin"
)

func SetupAPI() *gin.Engine {
	router := gin.Default()

	// manager := websocket.NewManager()

	// ma

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "ping"})
	})

	// router.GET("/ws", websocketManager.serveWS)

	return router
}
