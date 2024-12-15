package conn

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"main/app/config"
)

var Redis *redis.Redis

func InitRedis() {
	Redis = redis.MustNewRedis(config.Conf.RedisConf)
}
