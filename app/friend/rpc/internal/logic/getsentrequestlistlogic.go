package logic

import (
	"context"

	"SkyeIM/app/friend/rpc/friend"
	"SkyeIM/app/friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetSentRequestListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSentRequestListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSentRequestListLogic {
	return &GetSentRequestListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取发出的好友申请列表
func (l *GetSentRequestListLogic) GetSentRequestList(in *friend.GetSentRequestListReq) (*friend.GetSentRequestListResp, error) {
	// 默认分页参数
	page := in.Page
	pageSize := in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 查询发出的申请列表
	requests, err := l.svcCtx.FriendRequestModel.FindByFromUserId(l.ctx, uint64(in.UserId), page, pageSize)
	if err != nil {
		l.Logger.Errorf("查询发出的申请列表失败: %v", err)
		return nil, status.Error(codes.Internal, "查询申请列表失败")
	}

	// 统计总数
	total, err := l.svcCtx.FriendRequestModel.CountByFromUserId(l.ctx, uint64(in.UserId))
	if err != nil {
		l.Logger.Errorf("统计发出的申请数量失败: %v", err)
		return nil, status.Error(codes.Internal, "统计申请数量失败")
	}

	// 转换响应
	list := make([]*friend.FriendRequestInfo, 0, len(requests))
	for _, r := range requests {
		list = append(list, &friend.FriendRequestInfo{
			Id:         int64(r.Id),
			FromUserId: int64(r.FromUserId),
			ToUserId:   int64(r.ToUserId),
			Message:    r.Message,
			Status:     r.Status,
			CreatedAt:  r.CreatedAt.Unix(),
		})
	}

	return &friend.GetSentRequestListResp{
		List:  list,
		Total: total,
	}, nil
}
