// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package membermgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"SkyeIM/app/group/api/internal/svc"
	"SkyeIM/app/group/api/internal/types"
	"SkyeIM/app/group/rpc/groupclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMemberListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetMemberListLogic creates a new GetMemberListLogic.
func NewGetMemberListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMemberListLogic {
	return &GetMemberListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMemberListLogic) GetMemberList(req *types.GetMemberListReq) (resp *types.Response, err error) {
	userId := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	uid, _ := userId.Int64()

	rpcRes, err := l.svcCtx.GroupRpc.GetMemberList(l.ctx, &groupclient.GetMemberListReq{
		GroupId:  req.GroupId,
		UserId:   uid,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.MemberInfo, 0, len(rpcRes.Members))
	for _, v := range rpcRes.Members {
		joinedAt := time.Unix(v.JoinedAt, 0).Format("2006-01-02 15:04:05")
		list = append(list, types.MemberInfo{
			UserId:   v.UserId,
			Nickname: v.Nickname,
			Avatar:   "",
			Role:     v.Role,
			Mute:     v.Mute,
			JoinTime: v.JoinedAt,
			JoinedAt: joinedAt,
			ReadSeq:  v.ReadSeq,
		})
	}

	return &types.Response{
		Code:    0,
		Message: "success",
		Data: types.GetMemberListResp{
			List:  list,
			Total: int64(rpcRes.Total),
		},
	}, nil
}
