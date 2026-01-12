// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package groupmgmt

import (
	"context"
	"encoding/json"
	"fmt"

	"SkyeIM/app/group/api/internal/svc"
	"SkyeIM/app/group/api/internal/types"
	"SkyeIM/app/group/rpc/groupclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetGroupListLogic creates a new GetGroupListLogic.
func NewGetGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupListLogic {
	return &GetGroupListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupListLogic) GetGroupList(req *types.GetGroupListReq) (resp *types.Response, err error) {
	userId := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	uid, _ := userId.Int64()

	rpcRes, err := l.svcCtx.GroupRpc.GetUserGroupList(l.ctx, &groupclient.GetUserGroupListReq{
		UserId:   uid,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.GroupInfo, 0, len(rpcRes.Groups))
	for _, v := range rpcRes.Groups {
		list = append(list, types.GroupInfo{
			GroupId:     v.GroupId,
			Name:        v.Name,
			Avatar:      v.Avatar,
			OwnerId:     v.OwnerId,
			Description: v.Description,
			MaxMembers:  v.MaxMembers,
			MemberCount: v.MemberCount,
			Status:      v.Status,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	return &types.Response{
		Code:    0,
		Message: "success",
		Data: types.GetGroupListResp{
			List:  list,
			Total: int64(rpcRes.Total),
		},
	}, nil
}
