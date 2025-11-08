package conn

import (
	"github.com/limingxinleo/go-zero-skeleton/app/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var Redis *redis.Redis

func InitRedis() {
	Redis = redis.MustNewRedis(config.Conf.RedisConf)
}
