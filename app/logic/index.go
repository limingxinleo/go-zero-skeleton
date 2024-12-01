package logic

import (
	"context"
	"main/app/kernel"
	"main/app/svc"
	"main/app/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IndexService struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IndexService {
	return &IndexService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IndexService) Index(req *types.FromRequest) (result string, err kernel.ErrorCodeInterface) {
	result = req.Name
	return
}
