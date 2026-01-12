// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package message

import (
	"context"
	"encoding/json"

	"SkyeIM/app/message/api/internal/svc"
	"SkyeIM/app/message/api/internal/types"
	"SkyeIM/app/message/rpc/messageclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 模糊搜索聊天记录
func NewSearchMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchMessageLogic {
	return &SearchMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchMessageLogic) SearchMessage(req *types.SearchMessageReq) (resp *types.SearchMessageResp, err error) {
	userId, _ := l.ctx.Value("userId").(json.Number).Int64()

	rpcRes, err := l.svcCtx.MessageRpc.SearchMessage(l.ctx, &messageclient.SearchMessageReq{
		UserId:  userId,
		Keyword: req.Keyword,
	})
	if err != nil {
		l.Logger.Errorf("API 搜索消息失败: %v", err)
		return nil, err
	}

	var list []types.MessageInfo
	for _, v := range rpcRes.List {
		list = append(list, types.MessageInfo{
			Id:          v.Id,
			MsgId:       v.MsgId,
			FromUserId:  v.FromUserId,
			ToUserId:    v.ToUserId,
			ChatType:    v.ChatType,
			GroupId:     v.GroupId,
			Content:     v.Content,
			ContentType: v.ContentType,
			Status:      v.Status,
			CreatedAt:   v.CreatedAt,
			Seq:         v.Seq,
		})
	}

	return &types.SearchMessageResp{
		List: list,
	}, nil
}
