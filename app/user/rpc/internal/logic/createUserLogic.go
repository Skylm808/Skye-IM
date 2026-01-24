package logic

import (
	"context"
	"database/sql"
	"fmt"

	"SkyeIM/app/user/rpc/internal/svc"
	"SkyeIM/app/user/rpc/user"
	"auth/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建用户（注册用）
func (l *CreateUserLogic) CreateUser(in *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	// 设置昵称
	nickname := in.Nickname
	if nickname == "" {
		nickname = in.Username
	}

	// 创建用户
	newUser := &model.User{
		Username: in.Username,
		Password: in.Password, // 已加密的密码
		Email:    sql.NullString{String: in.Email, Valid: in.Email != ""},
		Phone:    sql.NullString{String: in.Phone, Valid: in.Phone != ""},
		Nickname: nickname,
		Avatar:   "",
		Status:   1, // 默认正常状态
	}

	result, err := l.svcCtx.UserModel.Insert(l.ctx, newUser)
	if err != nil {
		l.Logger.Errorf("创建用户失败: %v", err)
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	// 获取新用户ID
	userId, err := result.LastInsertId()
	if err != nil {
		l.Logger.Errorf("获取用户ID失败: %v", err)
		return nil, fmt.Errorf("获取用户ID失败: %v", err)
	}

	// 查询完整的用户信息
	createdUser, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(userId))
	if err != nil {
		l.Logger.Errorf("查询新创建用户失败: %v", err)
		return nil, fmt.Errorf("查询新创建用户失败: %v", err)
	}

	return &user.CreateUserResponse{
		UserId: userId,
		User:   convertToUserInfo(createdUser),
	}, nil
}
