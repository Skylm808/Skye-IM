// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"encoding/json"

	"SkyeIM/app/user/rpc/userClient"
	"SkyeIM/common/errorx"
	"auth/internal/svc"
	"auth/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户信息
func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo(req *types.Empty) (resp *types.UserInfoResponse, err error) {
	// 从JWT中获取用户ID
	userIdValue := l.ctx.Value("userId")
	if userIdValue == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 处理userId类型
	var userId int64
	switch v := userIdValue.(type) {
	case json.Number:
		userId, _ = v.Int64()
	case float64:
		userId = int64(v)
	case int64:
		userId = v
	default:
		l.Logger.Errorf("userId类型错误: %T", userIdValue)
		return nil, errorx.ErrUnauthorized
	}

	// 通过RPC查询用户信息
	userResp, err := l.svcCtx.UserRpc.GetUser(l.ctx, &userClient.GetUserRequest{
		Id: userId,
	})
	if err != nil {
		l.Logger.Errorf("RPC查询用户失败: %v", err)
		return nil, errorx.ErrUserNotFound
	}

	// 检查用户状态
	if userResp.User.Status == 0 {
		return nil, errorx.ErrUserDisabled
	}

	return &types.UserInfoResponse{
		Id:       userResp.User.Id,
		Username: userResp.User.Username,
		Phone:    userResp.User.Phone,
		Email:    userResp.User.Email,
		Nickname: userResp.User.Nickname,
		Avatar:   userResp.User.Avatar,
		Status:   userResp.User.Status,
	}, nil
}
