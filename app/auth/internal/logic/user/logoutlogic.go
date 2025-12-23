// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"auth/internal/svc"
	"auth/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 退出登录
func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout(req *types.Empty) (resp *types.Empty, err error) {
	// 从JWT中获取用户ID
	userId := l.ctx.Value("userId")

	// 目前简单实现：客户端删除Token即可
	// 高级实现：可以将Token加入黑名单（存入Redis）
	// 在认证中间件中检查Token是否在黑名单中

	l.Logger.Infof("用户退出登录: userId=%v", userId)

	return &types.Empty{}, nil
}
