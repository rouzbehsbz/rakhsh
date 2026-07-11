package apiUtils

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetClientId(c *gin.Context) (int32, error) {
	value, ok := c.Get("client_id")
	if !ok {
		return 0, fmt.Errorf("can't find client id")
	}

	clientId, ok := value.(int32)
	if !ok {
		return 0, fmt.Errorf("wrong format of client id")
	}

	return clientId, nil
}

func SendSuccessJson(c *gin.Context, message string, result any) {
	response := NewSuccessResponse(message, result)

	c.JSON(response.GetHttpStatusCode(), response)
}

func SendError(c *gin.Context, err error) {
	c.Abort()
	c.Error(err)
}
