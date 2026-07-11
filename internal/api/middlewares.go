package api

import (
	apiUtils "rakhsh/internal/api/utils"
	"rakhsh/internal/common"

	"github.com/gin-gonic/gin"
)

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		appErr, ok := err.(*common.AppError)
		if !ok {
			appErr = common.InternalError("")
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
