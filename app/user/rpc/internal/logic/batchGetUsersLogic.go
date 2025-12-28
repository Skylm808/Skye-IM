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

type BatchGetUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetUsersLogic {
	return &BatchGetUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量获取用户信息（消息列表、群成员列表等场景）
func (l *BatchGetUsersLogic) BatchGetUsers(in *user.BatchGetUsersRequest) (*user.BatchGetUsersResponse, error) {
	// 参数校验
	if len(in.Ids) == 0 {
		return &user.BatchGetUsersResponse{Users: []*user.UserInfo{}}, nil
	}
	if len(in.Ids) > 100 {
		return nil, status.Error(codes.InvalidArgument, "批量查询最多支持100个用户")
	}

	// 批量查询用户
	users := make([]*user.UserInfo, 0, len(in.Ids))
	for _, id := range in.Ids {
		if id <= 0 {
			continue
		}
		userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(id))
		if err != nil {
			if err == model.ErrNotFound {
				continue // 跳过不存在的用户
			}
			l.Logger.Errorf("查询用户 %d 失败: %v", id, err)
			continue
		}
		users = append(users, convertToUserInfo(userInfo))
	}

	return &user.BatchGetUsersResponse{Users: users}, nil
}
