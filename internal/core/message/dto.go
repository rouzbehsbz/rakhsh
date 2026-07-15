package message

type PostMessageRequest struct {
	Recipients []string `json:"recipients" binding:"required"`
	Text       string   `json:"text" binding:"required"`
	IsExpress  *bool    `json:"isExpress" binding:"required"`
}

type GetReportsRequest struct {
	Uids []uint64 `form:"uids" binding:"required"`
}

type PostMessageInput struct {
	ClientId   int32
	Recipients []string
	Text       string
	IsExpress  bool
}

type PostMessageOutput struct {
	Uids []string `json:"uids"`
}

type GetReportsOutput struct {
	Messages []Message `json:"messages"`
}
