package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/message/model"
	"SkyeIM/app/message/rpc/internal/svc"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SendGroupMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendGroupMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendGroupMessageLogic {
	return &SendGroupMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送群聊消息
func (l *SendGroupMessageLogic) SendGroupMessage(in *message.SendGroupMessageReq) (*message.SendGroupMessageResp, error) {
	if in.MsgId == "" || in.FromUserId == 0 || in.GroupId == "" || in.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "参数错误")
	}

	checkResp, err := l.svcCtx.GroupRpc.CheckMembership(l.ctx, &group.CheckMembershipReq{
		GroupId: in.GroupId,
		UserId:  in.FromUserId,
	})
	if err != nil {
		l.Logger.Errorf("检查成员资格失败: %v", err)
		return nil, status.Error(codes.Internal, "检查成员失败")
	}

	if !checkResp.IsMember {
		return nil, status.Error(codes.PermissionDenied, "您不是群成员")
	}

	if checkResp.Member.Mute == 1 {
		return nil, status.Error(codes.PermissionDenied, "您已被禁言")
	}

	// 验证@权限：如果包含@全体(-1)，只有群主和管理员可以使用
	if containsAtAll(in.AtUserIds) {
		if checkResp.Member.Role != 1 && checkResp.Member.Role != 2 {
			return nil, status.Error(codes.PermissionDenied, "只有群主和管理员可以@全体成员")
		}
	}

	// 1. 生成 Seq
	seqKey := fmt.Sprintf("group:seq:%s", in.GroupId)
	seq, err := l.svcCtx.Redis.Incr(seqKey)
	if err != nil {
		l.Logger.Errorf("生成群消息 Seq 失败: %v", err)
		return nil, status.Error(codes.Internal, "系统错误")
	}

	contentType := in.ContentType
	if contentType == 0 {
		contentType = 1
	}

	// 将at_user_ids转换为JSON字符串存储
	var atUserIdsJSON sql.NullString
	if len(in.AtUserIds) > 0 {
		atUserIdsBytes, err := json.Marshal(in.AtUserIds)
		if err == nil {
			atUserIdsJSON = sql.NullString{String: string(atUserIdsBytes), Valid: true}
		}
	}

	msgData := &model.ImMessage{
		MsgId:       in.MsgId,
		FromUserId:  uint64(in.FromUserId),
		ChatType:    2,
		GroupId:     sql.NullString{String: in.GroupId, Valid: true},
		Seq:         uint64(seq),
		Content:     in.Content,
		ContentType: int64(contentType),
		Status:      0,
		AtUserIds:   atUserIdsJSON,
	}

	result, err := l.svcCtx.ImMessageModel.Insert(l.ctx, msgData)
	if err != nil {
		l.Logger.Errorf("插入群聊消息失败: %v", err)
		return nil, status.Error(codes.Internal, "发送消息失败")
	}

	msgId, _ := result.LastInsertId()
	inserted, _ := l.svcCtx.ImMessageModel.FindOne(l.ctx, uint64(msgId))

	return &message.SendGroupMessageResp{
		Id:        msgId,
		MsgId:     in.MsgId,
		CreatedAt: inserted.CreatedAt.Unix(),
		Seq:       inserted.Seq,
	}, nil
}

// containsAtAll 检查是否包含@全体标识
func containsAtAll(ids []int64) bool {
	for _, id := range ids {
		if id == -1 {
			return true
		}
	}
	return false
}
