package logic

import (
	"context"

	"SkyeIM/app/message/model"
	"SkyeIM/app/message/rpc/internal/svc"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送消息（存储到数据库）
func (l *SendMessageLogic) SendMessage(in *message.SendMessageReq) (*message.SendMessageResp, error) {
	// 创建消息记录
	msg := &model.ImMessage{
		MsgId:       in.MsgId,
		FromUserId:  uint64(in.FromUserId),
		ToUserId:    uint64(in.ToUserId),
		ChatType:    1,
		Content:     in.Content,
		ContentType: int64(in.ContentType),
		Status:      0, // 默认未读
	}

	// 插入数据库
	result, err := l.svcCtx.ImMessageModel.Insert(l.ctx, msg)
	if err != nil {
		l.Logger.Errorf("SendMessage Insert failed: %v", err)
		return nil, err
	}

	// 获取插入的ID
	id, err := result.LastInsertId()
	if err != nil {
		l.Logger.Errorf("SendMessage LastInsertId failed: %v", err)
		return nil, err
	}

	// 获取消息详情（包含服务器时间戳）
	inserted, err := l.svcCtx.ImMessageModel.FindOne(l.ctx, uint64(id))
	if err != nil {
		l.Logger.Errorf("SendMessage FindOne failed: %v", err)
		return nil, err
	}

	return &message.SendMessageResp{
		Id:        id,
		MsgId:     in.MsgId,
		CreatedAt: inserted.CreatedAt.Unix(),
	}, nil
}
