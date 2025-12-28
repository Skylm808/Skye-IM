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

type SetBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetBlacklistLogic {
	return &SetBlacklistLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置黑名单（拉黑/取消拉黑）
func (l *SetBlacklistLogic) SetBlacklist(in *friend.SetBlacklistReq) (*friend.SetBlacklistResp, error) {
	// 1. 查询好友关系
	friendRecord, err := l.svcCtx.FriendModel.FindOneByUserIdFriendId(l.ctx, uint64(in.UserId), uint64(in.FriendId))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(codes.NotFound, "你们不是好友")
		}
		return nil, status.Error(codes.Internal, "查询好友关系失败")
	}

	// 2. 设置状态
	if in.IsBlack {
		friendRecord.Status = 2 // 拉黑
	} else {
		friendRecord.Status = 1 // 正常
	}

	err = l.svcCtx.FriendModel.Update(l.ctx, friendRecord)
	if err != nil {
		l.Logger.Errorf("设置黑名单失败: %v", err)
		return nil, status.Error(codes.Internal, "设置黑名单失败")
	}

	return &friend.SetBlacklistResp{}, nil
}
