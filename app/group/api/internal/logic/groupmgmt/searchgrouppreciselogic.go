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

type SearchGroupPreciseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSearchGroupPreciseLogic creates a new SearchGroupPreciseLogic.
func NewSearchGroupPreciseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchGroupPreciseLogic {
	return &SearchGroupPreciseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchGroupPreciseLogic) SearchGroupPrecise(req *types.SearchGroupPreciseReq) (resp *types.Response, err error) {
	userID := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	uid, err := userID.Int64()
	if err != nil {
		return &types.Response{
			Code:    400,
			Message: "invalid user id",
		}, nil
	}

	var groupInfo *groupclient.GroupInfo

	rpcRes, err := l.svcCtx.GroupRpc.GetGroupInfo(l.ctx, &groupclient.GetGroupInfoReq{
		GroupId: req.GroupId,
		UserId:  uid,
	})
	if err == nil && rpcRes.Group != nil {
		groupInfo = rpcRes.Group
	} else {
		searchRes, searchErr := l.svcCtx.GroupRpc.SearchGroup(l.ctx, &groupclient.SearchGroupReq{
			Keyword: req.GroupId,
		})
		if searchErr == nil {
			for _, v := range searchRes.Groups {
				if v.Name == req.GroupId || v.GroupId == req.GroupId {
					groupInfo = v
					break
				}
			}
		}
	}

	if groupInfo == nil {
		return &types.Response{
			Code:    404,
			Message: "group not found",
		}, nil
	}

	return &types.Response{
		Code:    0,
		Message: "success",
		Data: types.GroupInfo{
			GroupId:     groupInfo.GroupId,
			Name:        groupInfo.Name,
			Avatar:      groupInfo.Avatar,
			OwnerId:     groupInfo.OwnerId,
			Description: groupInfo.Description,
			MaxMembers:  groupInfo.MaxMembers,
			MemberCount: groupInfo.MemberCount,
			Status:      groupInfo.Status,
			CreatedAt:   groupInfo.CreatedAt,
			UpdatedAt:   groupInfo.UpdatedAt,
		},
	}, nil
}
