package redis

import (
	"context"

	"github.com/hawkneo/utils/web/health"
	"github.com/redis/go-redis/v9"
)

var (
	_ health.Indicator = (*RedisIndicator)(nil)
)

// NewRedisIndicator returns an Indicator based on go-redis package.
func NewRedisIndicator(name string, client redis.Cmdable) health.Indicator {
	return &RedisIndicator{
		name: name,
		cli:  client,
	}
}

type RedisIndicator struct {
	name string
	cli  redis.Cmdable
}

func (r RedisIndicator) Name() string {
	return r.name
}

func (r RedisIndicator) Health() health.Health {
	if err := r.cli.Info(context.TODO()).Err(); err != nil {
		return health.NewDownHealth(err)
	}
	return health.NewUpHealth()
}
