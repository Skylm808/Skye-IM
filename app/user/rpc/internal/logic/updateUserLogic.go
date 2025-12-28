package logic

import (
	"context"
	"database/sql"

	"SkyeIM/app/user/rpc/internal/svc"
	"SkyeIM/app/user/rpc/user"
	"auth/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户信息
func (l *UpdateUserLogic) UpdateUser(in *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	if in.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "无效的用户ID")
	}

	// 查询原用户信息
	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(codes.NotFound, "用户不存在")
		}
		l.Logger.Errorf("查询用户失败: %v", err)
		return nil, status.Error(codes.Internal, "查询用户失败")
	}

	// 更新字段（空字符串表示不更新）
	if in.Nickname != "" {
		userInfo.Nickname = in.Nickname
	}
	if in.Avatar != "" {
		userInfo.Avatar = in.Avatar
	}
	if in.Phone != "" {
		userInfo.Phone = sql.NullString{String: in.Phone, Valid: true}
	}

	// 保存更新
	err = l.svcCtx.UserModel.Update(l.ctx, userInfo)
	if err != nil {
		l.Logger.Errorf("更新用户失败: %v", err)
		return nil, status.Error(codes.Internal, "更新用户失败")
	}

	return &user.UpdateUserResponse{
		User: convertToUserInfo(userInfo),
	}, nil
}
