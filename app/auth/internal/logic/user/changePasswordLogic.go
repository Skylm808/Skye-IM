// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"encoding/json"

	"SkyeIM/common/errorx"
	"SkyeIM/common/utils"
	"auth/internal/svc"
	"auth/internal/types"
	"auth/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 修改密码
func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordRequest) (resp *types.ChangePasswordResponse, err error) {
	// 1. 参数校验
	if err := l.svcCtx.Validator.StructCtx(l.ctx, req); err != nil {
		return nil, errorx.NewCodeError(errorx.CodeParam, err.Error())
	}

	// 2. 从JWT中获取用户ID
	userIdValue := l.ctx.Value("userId")
	if userIdValue == nil {
		return nil, errorx.ErrUnauthorized
	}

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

	// 3. 查询用户
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(userId))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.ErrUserNotFound
		}
		l.Logger.Errorf("查询用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	// 4. 检查用户状态
	if user.Status == 0 {
		return nil, errorx.ErrUserDisabled
	}

	// 5. 验证旧密码
	if !utils.CheckPassword(req.OldPassword, user.Password) {
		return nil, errorx.NewCodeError(errorx.CodeParam, "旧密码错误")
	}

	// 6. 检查新旧密码是否相同
	if req.OldPassword == req.NewPassword {
		return nil, errorx.NewCodeError(errorx.CodeParam, "新密码不能与旧密码相同")
	}

	// 7. 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		l.Logger.Errorf("密码加密失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	// 8. 更新密码
	user.Password = hashedPassword
	if err := l.svcCtx.UserModel.Update(l.ctx, user); err != nil {
		l.Logger.Errorf("更新密码失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "修改密码失败")
	}

	l.Logger.Infof("密码修改成功: userId=%d", userId)

	return &types.ChangePasswordResponse{
		Message: "密码修改成功",
	}, nil
}
