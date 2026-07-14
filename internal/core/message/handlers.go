package message

import (
	apiUtils "rakhsh/internal/api/utils"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	service *MessageService
}

func NewMessageHandler(service *MessageService) *MessageHandler {
	return &MessageHandler{
		service: service,
	}
}

func (m *MessageHandler) PostMessage(c *gin.Context) {
	ctx := c.Request.Context()

	clientId, err := apiUtils.GetClientId(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req PostMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiUtils.SendError(c, err)
		return
	}

	output, err := m.service.PostMessage(ctx, PostMessageInput{
		ClientId:  clientId,
		Recipient: req.Recipient,
		Text:      req.Text,
		IsExpress: *req.IsExpress,
	})
	if err != nil {
		apiUtils.SendError(c, err)
		return
	}

	apiUtils.SendSuccessJson(c, "", output)
}

func (m *MessageHandler) GetReports(c *gin.Context) {
	ctx := c.Request.Context()

	clientId, err := apiUtils.GetClientId(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req GetReportsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apiUtils.SendError(c, err)
		return
	}

	output, err := m.service.GetReports(ctx, clientId, req.Uids)
	if err != nil {
		apiUtils.SendError(c, err)
		return
	}

	apiUtils.SendSuccessJson(c, "", output)
}
