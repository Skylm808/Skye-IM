package logic

import (
	"context"

	"SkyeIM/app/friend/rpc/friend"
	"SkyeIM/app/friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlacklistLogic {
	return &GetBlacklistLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取黑名单列表
func (l *GetBlacklistLogic) GetBlacklist(in *friend.GetBlacklistReq) (*friend.GetBlacklistResp, error) {
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

	// 查询黑名单列表
	blacklist, err := l.svcCtx.FriendModel.FindBlacklist(l.ctx, uint64(in.UserId), page, pageSize)
	if err != nil {
		l.Logger.Errorf("查询黑名单失败: %v", err)
		return nil, status.Error(codes.Internal, "查询黑名单失败")
	}

	// 统计总数
	total, err := l.svcCtx.FriendModel.CountBlacklist(l.ctx, uint64(in.UserId))
	if err != nil {
		l.Logger.Errorf("统计黑名单数量失败: %v", err)
		return nil, status.Error(codes.Internal, "统计黑名单数量失败")
	}

	// 转换响应
	list := make([]*friend.FriendInfo, 0, len(blacklist))
	for _, f := range blacklist {
		list = append(list, &friend.FriendInfo{
			Id:        int64(f.Id),
			FriendId:  int64(f.FriendId),
			Remark:    f.Remark,
			Status:    f.Status,
			CreatedAt: f.CreatedAt.Unix(),
		})
	}

	return &friend.GetBlacklistResp{
		List:  list,
		Total: total,
	}, nil
}
