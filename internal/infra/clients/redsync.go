package clients

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
)

func NewRedSync(
	data *DataClients,
) *redsync.Redsync {
	client := goredis.NewPool(data.RDB)
	return redsync.New(client)
}
