package logic

import (
	"context"
	"fmt"

	"SkyeIM/app/user/rpc/internal/svc"
	"SkyeIM/app/user/rpc/user"
	"auth/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePasswordLogic {
	return &UpdatePasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户密码（忘记密码、修改密码用）
func (l *UpdatePasswordLogic) UpdatePassword(in *user.UpdatePasswordRequest) (*user.UpdatePasswordResponse, error) {
	if in.UserId <= 0 {
		return &user.UpdatePasswordResponse{Success: false}, fmt.Errorf("无效的用户ID")
	}

	// 查询用户
	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if err != nil {
		if err == model.ErrNotFound {
			return &user.UpdatePasswordResponse{Success: false}, fmt.Errorf("用户不存在")
		}
		l.Logger.Errorf("查询用户失败: %v", err)
		return &user.UpdatePasswordResponse{Success: false}, err
	}

	// 更新密码
	userInfo.Password = in.NewPassword

	// 保存更新
	err = l.svcCtx.UserModel.Update(l.ctx, userInfo)
	if err != nil {
		l.Logger.Errorf("更新密码失败: %v", err)
		return &user.UpdatePasswordResponse{Success: false}, err
	}

	return &user.UpdatePasswordResponse{Success: true}, nil
}
