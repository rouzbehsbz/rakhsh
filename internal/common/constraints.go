package common

import "github.com/shopspring/decimal"

const AuthorizationHeaderPrefix = "JustId "
const PublishRequestBufferedChannelSize = 1024
const PendingMessagesQueueName = "pending_messages"

var PostMessageCost, _ = decimal.NewFromString("500")
