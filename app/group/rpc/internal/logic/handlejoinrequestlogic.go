package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"SkyeIM/app/group/model"
	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleJoinRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleJoinRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleJoinRequestLogic {
	return &HandleJoinRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HandleJoinRequest 处理入群申请（群主/管理员）
func (l *HandleJoinRequestLogic) HandleJoinRequest(in *group.HandleJoinRequestReq) (*group.HandleJoinRequestResp, error) {
	// 1. 查询申请记录
	request, err := l.svcCtx.ImGroupJoinRequestModel.FindOne(l.ctx, uint64(in.RequestId))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("申请记录不存在")
		}
		return nil, err
	}

	// 2. 检查申请状态（必须是待处理）
	if request.Status != 0 {
		return nil, errors.New("该申请已被处理")
	}

	// 3. 验证操作者权限（必须是群主或管理员）
	member, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, request.GroupId, in.OperatorId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("您不是群成员，无权处理申请")
		}
		return nil, err
	}
	if member.Role != 1 && member.Role != 2 {
		return nil, errors.New("只有群主或管理员可以处理申请")
	}

	// 4. 如果同意（action=1），添加用户到群组
	if in.Action == 1 {
		// 检查群组是否存在
		groupInfo, err := l.svcCtx.ImGroupModel.FindOneByGroupId(l.ctx, request.GroupId)
		if err != nil {
			return nil, err
		}
		if groupInfo.Status != 1 {
			return nil, errors.New("群组已解散")
		}

		// 检查群成员是否已满
		if groupInfo.MemberCount >= groupInfo.MaxMembers {
			return nil, errors.New("群成员已满，无法同意申请")
		}

		// 检查是否已经是成员（防止并发问题）
		existingMember, _ := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, request.GroupId, int64(request.UserId))
		if existingMember == nil {
			// 添加成员到群组
			newMember := &model.ImGroupMember{
				GroupId:  request.GroupId,
				UserId:   int64(request.UserId),
				Role:     3, // 普通成员
				Mute:     0, // 不禁言
				JoinedAt: time.Now(),
				ReadSeq:  0, // 初始已读序列号为0
			}
			_, err = l.svcCtx.ImGroupMemberModel.Insert(l.ctx, newMember)
			if err != nil {
				return nil, err
			}

			// 更新群组成员数
			groupInfo.MemberCount++
			err = l.svcCtx.ImGroupModel.Update(l.ctx, groupInfo)
			if err != nil {
				l.Logger.Errorf("更新群组成员数失败: %v", err)
				// 不中断流程，继续处理申请
			}
		}
	}

	// 5. 更新申请状态和处理人
	request.Status = in.Action // 1-同意, 2-拒绝
	request.HandlerId = sql.NullInt64{
		Int64: in.OperatorId,
		Valid: true,
	}
	err = l.svcCtx.ImGroupJoinRequestModel.Update(l.ctx, request)
	if err != nil {
		return nil, err
	}

	return &group.HandleJoinRequestResp{
		Success: true,
	}, nil
}
