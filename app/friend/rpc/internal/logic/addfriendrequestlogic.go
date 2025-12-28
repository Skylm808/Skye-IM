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

type AddFriendRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddFriendRequestLogic {
	return &AddFriendRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ==================== 好友申请相关 ====================
func (l *AddFriendRequestLogic) AddFriendRequest(in *friend.AddFriendRequestReq) (*friend.AddFriendRequestResp, error) {
	// 1. 不能添加自己
	if in.FromUserId == in.ToUserId {
		return nil, status.Error(codes.InvalidArgument, "不能添加自己为好友")
	}

	// 2. 检查是否已经是好友
	existFriend, err := l.svcCtx.FriendModel.FindOneByUserIdFriendId(l.ctx, uint64(in.FromUserId), uint64(in.ToUserId))
	if err != nil && err != model.ErrNotFound {
		return nil, status.Error(codes.Internal, "查询好友关系失败")
	}
	if existFriend != nil {
		return nil, status.Error(codes.AlreadyExists, "你们已经是好友了")
	}

	// 3. 检查是否有待处理的申请
	pendingRequest, err := l.svcCtx.FriendRequestModel.FindPendingRequest(l.ctx, uint64(in.FromUserId), uint64(in.ToUserId))
	if err != nil && err != model.ErrNotFound {
		return nil, status.Error(codes.Internal, "查询申请记录失败")
	}
	if pendingRequest != nil {
		return nil, status.Error(codes.AlreadyExists, "已有待处理的好友申请")
	}

	// 4. 检查对方是否已经发送申请给我
	reversePending, err := l.svcCtx.FriendRequestModel.FindPendingRequest(l.ctx, uint64(in.ToUserId), uint64(in.FromUserId))
	if err != nil && err != model.ErrNotFound {
		return nil, status.Error(codes.Internal, "查询申请记录失败")
	}
	if reversePending != nil {
		return nil, status.Error(codes.AlreadyExists, "对方已向你发送好友申请，请前往处理")
	}

	// 5. 创建好友申请
	result, err := l.svcCtx.FriendRequestModel.Insert(l.ctx, &model.ImFriendRequest{
		FromUserId: uint64(in.FromUserId),
		ToUserId:   uint64(in.ToUserId),
		Message:    in.Message,
		Status:     0, // 待处理
	})
	if err != nil {
		l.Logger.Errorf("创建好友申请失败: %v", err)
		return nil, status.Error(codes.Internal, "发送好友申请失败")
	}

	requestId, _ := result.LastInsertId()
	return &friend.AddFriendRequestResp{
		RequestId: requestId,
	}, nil
}
