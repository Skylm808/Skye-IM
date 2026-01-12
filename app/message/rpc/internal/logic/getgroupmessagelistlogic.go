package logic

import (
	"context"
	"encoding/json"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/message/rpc/internal/svc"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetGroupMessageListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupMessageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupMessageListLogic {
	return &GetGroupMessageListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取群聊历史消息列表
func (l *GetGroupMessageListLogic) GetGroupMessageList(in *message.GetGroupMessageListReq) (*message.GetGroupMessageListResp, error) {
	if in.GroupId == "" {
		return nil, status.Error(codes.InvalidArgument, "群组ID不能为空")
	}

	if in.UserId > 0 {
		checkResp, err := l.svcCtx.GroupRpc.CheckMembership(l.ctx, &group.CheckMembershipReq{
			GroupId: in.GroupId,
			UserId:  in.UserId,
		})
		if err != nil {
			l.Logger.Errorf("检查成员资格失败: %v", err)
			return nil, status.Error(codes.Internal, "检查成员失败")
		}

		if !checkResp.IsMember {
			return nil, status.Error(codes.PermissionDenied, "您不是群成员")
		}
	}

	limit := in.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	messages, err := l.svcCtx.ImMessageModel.FindGroupMessageList(l.ctx, in.GroupId, in.LastMsgId, int64(limit)+1)
	if err != nil {
		l.Logger.Errorf("查询群聊消息失败: %v", err)
		return nil, status.Error(codes.Internal, "查询消息失败")
	}

	hasMore := false
	if int64(len(messages)) > int64(limit) {
		hasMore = true
		messages = messages[:limit]
	}

	var list []*message.MessageInfo
	for _, msg := range messages {
		// 解析at_user_ids JSON
		var atUserIds []int64
		if msg.AtUserIds.Valid && msg.AtUserIds.String != "" {
			_ = json.Unmarshal([]byte(msg.AtUserIds.String), &atUserIds)
		}

		list = append(list, &message.MessageInfo{
			Id:          int64(msg.Id),
			MsgId:       msg.MsgId,
			FromUserId:  int64(msg.FromUserId),
			ToUserId:    int64(msg.ToUserId),
			ChatType:    int32(msg.ChatType),
			GroupId:     msg.GroupId.String,
			Content:     msg.Content,
			ContentType: int32(msg.ContentType),
			Status:      int32(msg.Status),
			CreatedAt:   msg.CreatedAt.Unix(),
			Seq:         msg.Seq,
			AtUserIds:   atUserIds,
		})
	}

	return &message.GetGroupMessageListResp{
		List:    list,
		HasMore: hasMore,
	}, nil
}
