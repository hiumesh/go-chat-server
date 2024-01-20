package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiumesh/go-chat-server/internal/conf"
	"github.com/hiumesh/go-chat-server/internal/utils"
)

func addUniqueRequestID(globalConfig *conf.GlobalConfiguration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		uid := uuid.Must(uuid.NewV6())
		id := uid.String()

		utils.WithRequestID(ctx, id)

	}
}
