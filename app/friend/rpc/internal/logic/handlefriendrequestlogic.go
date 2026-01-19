package logic

import (
	"context"
	"time"

	"SkyeIM/app/friend/rpc/friend"
	"SkyeIM/app/friend/rpc/internal/svc"
	"SkyeIM/app/friend/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HandleFriendRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleFriendRequestLogic {
	return &HandleFriendRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 处理好友申请（同意/拒绝）
func (l *HandleFriendRequestLogic) HandleFriendRequest(in *friend.HandleFriendRequestReq) (*friend.HandleFriendRequestResp, error) {
	// 1. 查询申请记录
	request, err := l.svcCtx.FriendRequestModel.FindOne(l.ctx, uint64(in.RequestId))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(codes.NotFound, "申请记录不存在")
		}
		return nil, status.Error(codes.Internal, "查询申请记录失败")
	}

	// 2. 验证接收方
	if request.ToUserId != uint64(in.UserId) {
		return nil, status.Error(codes.PermissionDenied, "无权处理此申请")
	}

	// 3. 检查申请状态
	if request.Status != 0 {
		return nil, status.Error(codes.FailedPrecondition, "申请已被处理")
	}

	// 4. 处理申请
	if in.Action == 1 {
		// 同意 - 双向建立好友关系
		_, err = l.svcCtx.FriendModel.Insert(l.ctx, &model.ImFriend{
			UserId:   request.FromUserId,
			FriendId: request.ToUserId,
			Status:   1,
		})
		if err != nil {
			l.Logger.Errorf("建立好友关系失败(正向): %v", err)
			return nil, status.Error(codes.Internal, "处理申请失败")
		}

		_, err = l.svcCtx.FriendModel.Insert(l.ctx, &model.ImFriend{
			UserId:   request.ToUserId,
			FriendId: request.FromUserId,
			Status:   1,
		})
		if err != nil {
			l.Logger.Errorf("建立好友关系失败(反向): %v", err)
			return nil, status.Error(codes.Internal, "处理申请失败")
		}

		// 更新申请状态为已同意
		err = l.svcCtx.FriendRequestModel.UpdateStatus(l.ctx, uint64(in.RequestId), 1)
	} else {
		// 拒绝 - 更新申请状态为已拒绝
		err = l.svcCtx.FriendRequestModel.UpdateStatus(l.ctx, uint64(in.RequestId), 2)
	}

	if err != nil {
		l.Logger.Errorf("更新申请状态失败: %v", err)
		return nil, status.Error(codes.Internal, "处理申请失败")
	}

	// 5. 推送通知给请求方
	go func() {
		actionText := "rejected"
		if in.Action == 1 {
			actionText = "accepted"
		}
		_ = l.svcCtx.WsPushClient.PushToUser(int64(request.FromUserId), "friend_request_handled", map[string]interface{}{
			"requestId": in.RequestId,
			"toUserId":  request.ToUserId,
			"action":    actionText,
			"handledAt": time.Now().Unix(),
		})
	}()

	return &friend.HandleFriendRequestResp{}, nil
}
