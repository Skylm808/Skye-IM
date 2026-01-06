// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package message

import (
	"context"

	"SkyeIM/app/message/api/internal/svc"
	"SkyeIM/app/message/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取会话列表（最近联系人）
func NewGetConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationsLogic {
	return &GetConversationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConversationsLogic) GetConversations(req *types.Empty) (resp *types.GetConversationsResp, err error) {
	// 从 JWT 获取当前用户ID
	_, err = getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// TODO: 实现获取会话列表
	// 这需要在 Message RPC 中添加新方法，暂时返回空列表
	// 后续可以添加 GetConversations RPC 方法来获取最近联系人

	return &types.GetConversationsResp{
		List: []types.ConversationInfo{},
	}, nil
}
