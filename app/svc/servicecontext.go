package svc

import (
	"github.com/limingxinleo/go-zero-skeleton/app/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
