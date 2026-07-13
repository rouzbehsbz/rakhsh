package message

type PostMessageRequest struct {
	Recipient string `json:"recipient" binding:"required"`
	Text      string `json:"text" binding:"required"`
	IsExpress *bool  `json:"isExpress" binding:"required"`
}

type PostMessageInput struct {
	ClientId  int32
	Recipient string
	Text      string
	IsExpress bool
}

type PostMessageOutput struct {
	Uid string `json:"uid"`
}

type GetMessageOutput struct {
	Messages []Message `json:"message"`
}
