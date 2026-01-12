package logic

import (
	"context"
	"errors"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateGroupReadSeqLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateGroupReadSeqLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGroupReadSeqLogic {
	return &UpdateGroupReadSeqLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateGroupReadSeq 更新群读消息位置
func (l *UpdateGroupReadSeqLogic) UpdateGroupReadSeq(in *group.UpdateGroupReadSeqReq) (*group.UpdateGroupReadSeqResp, error) {
	if in.GroupId == "" || in.UserId == 0 {
		return nil, errors.New("参数错误")
	}

	// 调用 model 中的 UpdateReadSeq 方法
	err := l.svcCtx.ImGroupMemberModel.UpdateReadSeq(l.ctx, in.GroupId, in.UserId, uint64(in.ReadSeq))
	if err != nil {
		l.Logger.Errorf("更新群读消息位置失败: %v", err)
		return nil, errors.New("更新失败")
	}

	return &group.UpdateGroupReadSeqResp{}, nil
}
