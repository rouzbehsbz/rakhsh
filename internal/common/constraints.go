package common

import "github.com/shopspring/decimal"

const PendingMessagesQueueName = "pending_messages"
const SubmittedMessagesQueueName = "submitted_messages"
const DeliveredMessageQueueName = "delivered_messages"
const RejectedMessagesQueueName = "rejected_messages"

// TODO: need to fetch from database
var PostMessageCost, _ = decimal.NewFromString("500")
