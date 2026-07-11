package apiUtils

import (
	"net/http"
	"rakhsh/pkg/utils"
	"time"
)

type Response struct {
	IsError   bool    `json:"isError"`
	Message   *string `json:"message"`
	Result    any     `json:"result"`
	Timestamp int64   `json:"timestamp"`

	httpStatusCode int
}

func NewSuccessResponse(message string, result any) Response {
	return Response{
		IsError:        false,
		Message:        utils.PtrString(message),
		Result:         result,
		Timestamp:      time.Now().UnixMilli(),
		httpStatusCode: http.StatusOK,
	}
}

func NewErrorResponse(httpStatusCode int, message string) Response {
	return Response{
		IsError:        true,
		Message:        utils.PtrString(message),
		Result:         nil,
		Timestamp:      time.Now().UnixMilli(),
		httpStatusCode: httpStatusCode,
	}
}

func (r Response) GetHttpStatusCode() int {
	return r.httpStatusCode
}
