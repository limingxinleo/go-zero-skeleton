package logic

import (
	"context"
	"main/internal/svc"
	"main/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MainLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MainLogic {
	return &MainLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MainLogic) Main(req *types.FromRequest) (resp *types.Response[string], err error) {
	resp = &types.Response[string]{
		Data: req.Name,
	}
	return
}
