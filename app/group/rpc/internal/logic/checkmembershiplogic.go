package logic

import (
	"context"

	"SkyeIM/app/group/model"
	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CheckMembershipLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckMembershipLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckMembershipLogic {
	return &CheckMembershipLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 检查成员资格
func (l *CheckMembershipLogic) CheckMembership(in *group.CheckMembershipReq) (*group.CheckMembershipResp, error) {
	if in.GroupId == "" || in.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "参数错误")
	}

	// 查询成员
	member, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.UserId)
	if err != nil {
		if err == model.ErrNotFound {
			return &group.CheckMembershipResp{
				IsMember: false,
			}, nil
		}
		l.Logger.Errorf("查询成员失败: %v", err)
		return nil, status.Error(codes.Internal, "查询失败")
	}

	return &group.CheckMembershipResp{
		IsMember: true,
		Member: &group.MemberInfo{
			Id:       member.Id,
			GroupId:  member.GroupId,
			UserId:   member.UserId,
			Role:     int32(member.Role),
			Nickname: member.Nickname.String,
			Mute:     int32(member.Mute),
			JoinedAt: member.JoinedAt.Unix(),
			ReadSeq:  member.ReadSeq,
		},
	}, nil
}
