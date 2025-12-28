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

type UpdateFriendRemarkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateFriendRemarkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFriendRemarkLogic {
	return &UpdateFriendRemarkLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新好友备注
func (l *UpdateFriendRemarkLogic) UpdateFriendRemark(in *friend.UpdateFriendRemarkReq) (*friend.UpdateFriendRemarkResp, error) {
	// 1. 查询好友关系
	friendRecord, err := l.svcCtx.FriendModel.FindOneByUserIdFriendId(l.ctx, uint64(in.UserId), uint64(in.FriendId))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(codes.NotFound, "你们不是好友")
		}
		return nil, status.Error(codes.Internal, "查询好友关系失败")
	}

	// 2. 更新备注
	friendRecord.Remark = in.Remark
	err = l.svcCtx.FriendModel.Update(l.ctx, friendRecord)
	if err != nil {
		l.Logger.Errorf("更新好友备注失败: %v", err)
		return nil, status.Error(codes.Internal, "更新备注失败")
	}

	return &friend.UpdateFriendRemarkResp{}, nil
}
