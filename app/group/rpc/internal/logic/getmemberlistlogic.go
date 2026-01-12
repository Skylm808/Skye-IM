package logic

import (
	"context"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMemberListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMemberListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMemberListLogic {
	return &GetMemberListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetMemberList 获取群成员列表
func (l *GetMemberListLogic) GetMemberList(in *group.GetMemberListReq) (*group.GetMemberListResp, error) {
	members, err := l.svcCtx.ImGroupMemberModel.FindByGroupId(l.ctx, in.GroupId)
	if err != nil {
		logx.Errorf("查询群成员失败: %v", err)
		return nil, err
	}

	var list []*group.MemberInfo
	for _, m := range members {
		list = append(list, &group.MemberInfo{
			Id:       m.Id,
			GroupId:  m.GroupId,
			UserId:   m.UserId,
			Role:     int32(m.Role),
			Nickname: m.Nickname.String,
			Mute:     int32(m.Mute),
			JoinedAt: m.JoinedAt.Unix(),
			ReadSeq:  m.ReadSeq,
		})
	}

	return &group.GetMemberListResp{
		Members: list,
		Total:   int32(len(list)),
	}, nil
}
