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

type GetGroupInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取群组详情
func NewGetGroupInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupInfoLogic {
	return &GetGroupInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupInfoLogic) GetGroupInfo(req *types.GetGroupInfoReq) (resp *types.Response, err error) {
	userID := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	uid, err := userID.Int64()
	if err != nil {
		return &types.Response{
			Code:    400,
			Message: "invalid user id",
		}, nil
	}

	rpcRes, err := l.svcCtx.GroupRpc.GetGroupInfo(l.ctx, &groupclient.GetGroupInfoReq{
		GroupId: req.GroupId,
		UserId:  uid,
	})

	if err != nil {
		return nil, err
	}

	return &types.Response{
		Code:    0,
		Message: "获取成功",
		Data: types.GroupInfo{
			GroupId:     rpcRes.Group.GroupId,
			Name:        rpcRes.Group.Name,
			Avatar:      rpcRes.Group.Avatar,
			OwnerId:     rpcRes.Group.OwnerId,
			Description: rpcRes.Group.Description,
			MaxMembers:  rpcRes.Group.MaxMembers,
			MemberCount: rpcRes.Group.MemberCount,
			Status:      rpcRes.Group.Status,
			CreatedAt:   rpcRes.Group.CreatedAt,
			UpdatedAt:   rpcRes.Group.UpdatedAt,
		},
	}, nil
}
