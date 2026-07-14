package message

type PostMessageRequest struct {
	Recipient string `json:"recipient" binding:"required"`
	Text      string `json:"text" binding:"required"`
	IsExpress *bool  `json:"isExpress" binding:"required"`
}

type GetReportsRequest struct {
	Uids []uint64 `form:"uids" binding:"required"`
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

type GetReportsOutput struct {
	Messages []Message `json:"messages"`
}
