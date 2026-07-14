package cronjob

import (
	"context"
	"time"
)

const BatchUpdatesSize = 100
const BatchUpdateInterval = 30 * time.Second

func main() {
	_ = context.Background()
}
