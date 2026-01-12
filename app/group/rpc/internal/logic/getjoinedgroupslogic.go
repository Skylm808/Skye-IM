package logic

import (
	"context"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetJoinedGroupsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetJoinedGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetJoinedGroupsLogic {
	return &GetJoinedGroupsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetJoinedGroups 获取加入的群组列表
func (l *GetJoinedGroupsLogic) GetJoinedGroups(in *group.GetJoinedGroupsReq) (*group.GetJoinedGroupsResp, error) {
	// 使用合理的默认分页参数获取所有群组
	// 1 表示第1页，1000 表示每页最多1000条（足够返回所有群组）
	members, err := l.svcCtx.ImGroupMemberModel.FindGroupsByUserId(l.ctx, in.UserId, 1, 1000)
	if err != nil {
		logx.Errorf("查询用户群组失败: %v", err)
		return nil, err
	}

	// 返回 MemberInfo 列表（包含 group_id, user_id, role, read_seq 等成员信息）
	var memberList []*group.MemberInfo
	for _, m := range members {
		memberList = append(memberList, &group.MemberInfo{
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

	return &group.GetJoinedGroupsResp{
		List: memberList,
	}, nil
}
