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

func (l *MainLogic) Main(req *types.Request) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
