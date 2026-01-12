// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package message

import (
	"context"

	"SkyeIM/app/message/api/internal/svc"
	"SkyeIM/app/message/api/internal/types"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupSyncLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 群聊离线同步（按seq拉取）
func NewGetGroupSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupSyncLogic {
	return &GetGroupSyncLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupSyncLogic) GetGroupSync(req *types.GetGroupSyncReq) (resp *types.GetGroupSyncResp, err error) {
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	if req.GroupId == "" {
		return &types.GetGroupSyncResp{List: []types.MessageInfo{}}, nil
	}

	rpcResp, err := l.svcCtx.MessageRpc.GetGroupMessagesBySeq(l.ctx, &message.GetGroupMessagesBySeqReq{
		UserId:  userId,
		GroupId: req.GroupId,
		Seq:     req.Seq,
	})
	if err != nil {
		l.Logger.Errorf("GetGroupSync RPC failed: %v", err)
		return nil, err
	}

	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 200
	}

	list := make([]types.MessageInfo, 0, len(rpcResp.List))
	for _, msg := range rpcResp.List {
		list = append(list, types.MessageInfo{
			Id:          msg.Id,
			MsgId:       msg.MsgId,
			FromUserId:  msg.FromUserId,
			ToUserId:    msg.ToUserId,
			ChatType:    msg.ChatType,
			GroupId:     msg.GroupId,
			Content:     msg.Content,
			ContentType: msg.ContentType,
			Status:      msg.Status,
			CreatedAt:   msg.CreatedAt,
			Seq:         msg.Seq,
		})
		if int32(len(list)) >= limit {
			break
		}
	}

	return &types.GetGroupSyncResp{List: list}, nil
}
