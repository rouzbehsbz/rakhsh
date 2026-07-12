package api

import (
	apiUtils "rakhsh/internal/api/utils"
	"rakhsh/internal/common"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			apiUtils.SendError(c, common.UnauthorizedError("Missing authorization header"))
			return
		}

		if !strings.HasPrefix(header, common.AuthorizationHeaderPrefix) {
			apiUtils.SendError(c, common.UnauthorizedError("Invalid authorization header"))
			return
		}

		sClientId := strings.TrimPrefix(header, common.AuthorizationHeaderPrefix)

		clientId, err := strconv.Atoi(sClientId)
		if err != nil {
			apiUtils.SendError(c, common.UnauthorizedError("Invalid authorization header"))
			return
		}

		c.Set("client_id", int32(clientId))

		c.Next()
	}
}

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		appErr, ok := err.(*common.AppError)
		if !ok {
			appErr = common.InternalError(err.Error())
		}

		response := apiUtils.NewErrorResponse(appErr.StatusCode, appErr.Message)

		c.AbortWithStatusJSON(response.GetHttpStatusCode(), response)
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		apiUtils.SendError(c, common.InternalError(""))
	})
}
