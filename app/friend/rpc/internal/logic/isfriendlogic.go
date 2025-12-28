package logic

import (
	"context"

	"SkyeIM/app/friend/rpc/friend"
	"SkyeIM/app/friend/rpc/internal/svc"
	"SkyeIM/app/friend/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IsFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsFriendLogic {
	return &IsFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 检查是否为好友
func (l *IsFriendLogic) IsFriend(in *friend.IsFriendReq) (*friend.IsFriendResp, error) {
	// 查询好友关系
	friendRecord, err := l.svcCtx.FriendModel.FindOneByUserIdFriendId(l.ctx, uint64(in.UserId), uint64(in.FriendId))
	if err != nil {
		if err == model.ErrNotFound {
			return &friend.IsFriendResp{
				IsFriend: false,
				Status:   0, // 非好友
			}, nil
		}
		l.Logger.Errorf("查询好友关系失败: %v", err)
		return nil, status.Error(codes.Internal, "查询好友关系失败")
	}

	return &friend.IsFriendResp{
		IsFriend: friendRecord.Status == 1, // status=1 为正常好友
		Status:   friendRecord.Status,
	}, nil
}
