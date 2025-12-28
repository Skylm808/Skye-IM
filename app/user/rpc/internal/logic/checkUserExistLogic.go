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

type CheckUserExistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckUserExistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckUserExistLogic {
	return &CheckUserExistLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 检查用户是否存在
func (l *CheckUserExistLogic) CheckUserExist(in *user.CheckUserExistRequest) (*user.CheckUserExistResponse, error) {
	if in.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "无效的用户ID")
	}

	_, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == model.ErrNotFound {
			return &user.CheckUserExistResponse{Exist: false}, nil
		}
		l.Logger.Errorf("查询用户失败: %v", err)
		return nil, status.Error(codes.Internal, "查询用户失败")
	}

	return &user.CheckUserExistResponse{Exist: true}, nil
}
