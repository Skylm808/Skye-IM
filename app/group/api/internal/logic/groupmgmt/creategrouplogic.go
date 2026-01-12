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

type CreateGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建群组
func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateGroupLogic) CreateGroup(req *types.CreateGroupReq) (resp *types.Response, err error) {
	userId := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	uid, _ := userId.Int64()

	rpcRes, err := l.svcCtx.GroupRpc.CreateGroup(l.ctx, &groupclient.CreateGroupReq{
		Name:        req.Name,
		Avatar:      req.Avatar,
		OwnerId:     uid,
		Description: req.Description,
		MaxMembers:  req.MaxMembers,
		MemberIds:   req.MemberIds,
	})

	if err != nil {
		return nil, err
	}

	return &types.Response{
		Code:    0,
		Message: "创建成功",
		Data: types.CreateGroupResp{
			GroupId: rpcRes.Group.GroupId,
		},
	}, nil
}
