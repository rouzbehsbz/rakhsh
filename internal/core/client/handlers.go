package client

import (
	apiUtils "rakhsh/internal/api/utils"

	"github.com/gin-gonic/gin"
)

type ClientHandler struct {
	service *ClientService
}

func NewClientHandler(service *ClientService) *ClientHandler {
	return &ClientHandler{
		service: service,
	}
}

func (ch *ClientHandler) GetSelfClientInfoHandler(c *gin.Context) {
	ctx := c.Request.Context()

	clientId, err := apiUtils.GetClientId(c)
	if err != nil {
		c.Error(err)
		return
	}

	output, err := ch.service.GetClientInfo(ctx, clientId)
	if err != nil {
		apiUtils.SendError(c, err)
		return
	}

	apiUtils.SendSuccessJson(c, "", output)
}

func (ch *ClientHandler) ChargeBalanceWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	clientId, err := apiUtils.GetClientId(c)
	if err != nil {
		return
	}

	var req ChargeBalanceWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiUtils.SendError(c, err)
		return
	}

	err = ch.service.ChargeBalance(ctx, clientId, req.Amount)
	if err != nil {
		apiUtils.SendError(c, err)
		return
	}

	apiUtils.SendSuccessJson(c, "account has been charged successfully.", nil)
}
