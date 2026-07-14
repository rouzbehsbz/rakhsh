package cronjob

import (
	"time"
)

const BatchUpdatesSize = 100
const BatchUpdateInterval = 30 * time.Second

func main() {
	//TODO: need a way to fetch pending messages and check if its delivered or not
}
