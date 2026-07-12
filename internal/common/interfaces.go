package common

import "context"

type Transactional interface {
	WithinTx(ctx context.Context, clientId int32, fn func(ctx context.Context) error) error
}
