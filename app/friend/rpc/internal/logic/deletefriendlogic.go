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

type DeleteFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFriendLogic {
	return &DeleteFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除好友
func (l *DeleteFriendLogic) DeleteFriend(in *friend.DeleteFriendReq) (*friend.DeleteFriendResp, error) {
	// 1. 检查是否是好友
	_, err := l.svcCtx.FriendModel.FindOneByUserIdFriendId(l.ctx, uint64(in.UserId), uint64(in.FriendId))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(codes.NotFound, "你们不是好友")
		}
		return nil, status.Error(codes.Internal, "查询好友关系失败")
	}

	// 2. 双向删除好友关系
	// 删除 我 -> 他
	err = l.svcCtx.FriendModel.DeleteByUserIdFriendId(l.ctx, uint64(in.UserId), uint64(in.FriendId))
	if err != nil {
		l.Logger.Errorf("删除好友关系失败(我->他): %v", err)
		return nil, status.Error(codes.Internal, "删除好友失败")
	}

	// 删除 他 -> 我
	err = l.svcCtx.FriendModel.DeleteByUserIdFriendId(l.ctx, uint64(in.FriendId), uint64(in.UserId))
	if err != nil {
		l.Logger.Errorf("删除好友关系失败(他->我): %v", err)
		return nil, status.Error(codes.Internal, "删除好友失败")
	}

	return &friend.DeleteFriendResp{}, nil
}
