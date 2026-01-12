package logic

import (
	"context"

	"SkyeIM/app/user/rpc/internal/svc"
	"SkyeIM/app/user/rpc/user"
	"auth/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个用户信息
func (l *GetUserLogic) GetUser(in *user.GetUserRequest) (*user.GetUserResponse, error) {
	// 参数校验
	if in.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "无效的用户ID")
	}

	// 查询用户
	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(codes.NotFound, "用户不存在")
		}
		l.Logger.Errorf("查询用户失败: %v", err)
		return nil, status.Error(codes.Internal, "查询用户失败")
	}

	return &user.GetUserResponse{
		User: convertToUserInfo(userInfo),
	}, nil
}

// convertToUserInfo 将 model.User 转换为 proto 的 UserInfo
func convertToUserInfo(u *model.User) *user.UserInfo {
	phone := ""
	if u.Phone.Valid {
		phone = u.Phone.String
	}
	email := ""
	if u.Email.Valid {
		email = u.Email.String
	}

	return &user.UserInfo{
		Id:        int64(u.Id),
		Username:  u.Username,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Phone:     phone,
		Email:     email,
		Signature: u.Signature,
		Gender:    int64(u.Gender),
		Region:    u.Region,
		Status:    int64(u.Status),
		CreatedAt: u.CreatedAt.Unix(),
	}
}
