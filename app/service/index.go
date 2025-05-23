package service

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"main/app/kernel"
	"main/app/svc"
	"main/app/types"
)

type IndexService struct {
	log    logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIndexService(ctx context.Context, svcCtx *svc.ServiceContext) *IndexService {
	return &IndexService{
		log:    logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IndexService) Index(req *types.FromRequest) (result string, err kernel.ErrorCodeInterface) {
	result = fmt.Sprintf("Hi %s, welcome to %s", req.Name, l.svcCtx.Config.Name)
	return
}
