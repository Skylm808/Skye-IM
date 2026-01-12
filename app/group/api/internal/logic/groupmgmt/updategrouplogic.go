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

type UpdateGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新群组信息
func NewUpdateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGroupLogic {
	return &UpdateGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateGroupLogic) UpdateGroup(req *types.UpdateGroupReq) (resp *types.Response, err error) {
	userId := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	uid, _ := userId.Int64()

	_, err = l.svcCtx.GroupRpc.UpdateGroup(l.ctx, &groupclient.UpdateGroupReq{
		GroupId:     req.GroupId,
		OperatorId:  uid,
		Name:        req.Name,
		Avatar:      req.Avatar,
		Description: req.Description,
		MaxMembers:  req.MaxMembers,
	})

	if err != nil {
		return nil, err
	}

	return &types.Response{
		Code:    0,
		Message: "更新成功",
	}, nil
}
