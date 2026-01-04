package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Environment string
	RedisConf   redis.RedisConf
	MySqlConf   MySqlConf
}

type MySqlConf struct {
	Dsn string `json:"dsn"`
}

func (c *Config) IsProd() bool {
	return c.Environment == "prod"
}
