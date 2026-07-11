package client

import "github.com/shopspring/decimal"

type ChargeBalanceWebhookRequest struct {
	Amount decimal.Decimal `json:"amount" binding:"required"`
}

type GetClientInfoOutput struct {
	Name    string          `json:"name"`
	Balance decimal.Decimal `json:"balance"`
}
