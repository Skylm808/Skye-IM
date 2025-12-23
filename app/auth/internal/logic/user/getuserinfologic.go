// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"encoding/json"

	"SkyeIM/common/errorx"
	"auth/internal/svc"
	"auth/internal/types"
	"auth/model"

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

	// 查询用户信息（注意：FindOne 需要 uint64）
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(userId))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.ErrUserNotFound
		}
		l.Logger.Errorf("查询用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	// 检查用户状态
	if user.Status == 0 {
		return nil, errorx.ErrUserDisabled
	}

	// 处理可空字段
	phone := ""
	if user.Phone.Valid {
		phone = user.Phone.String
	}
	email := ""
	if user.Email.Valid {
		email = user.Email.String
	}

	return &types.UserInfoResponse{
		Id:       int64(user.Id), // uint64 转 int64
		Username: user.Username,
		Phone:    phone,
		Email:    email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Status:   int64(user.Status), // uint64 转 int64
	}, nil
}
