package logic

import (
	"context"

	"SkyeIM/app/friend/rpc/friend"
	"SkyeIM/app/friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetFriendListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendListLogic {
	return &GetFriendListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取好友列表
func (l *GetFriendListLogic) GetFriendList(in *friend.GetFriendListReq) (*friend.GetFriendListResp, error) {
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

	// 查询好友列表（status=1 表示正常好友）
	friends, err := l.svcCtx.FriendModel.FindByUserId(l.ctx, uint64(in.UserId), 1, page, pageSize)
	if err != nil {
		l.Logger.Errorf("查询好友列表失败: %v", err)
		return nil, status.Error(codes.Internal, "查询好友列表失败")
	}

	// 统计总数
	total, err := l.svcCtx.FriendModel.CountByUserId(l.ctx, uint64(in.UserId), 1)
	if err != nil {
		l.Logger.Errorf("统计好友数量失败: %v", err)
		return nil, status.Error(codes.Internal, "统计好友数量失败")
	}

	// 转换响应
	list := make([]*friend.FriendInfo, 0, len(friends))
	for _, f := range friends {
		list = append(list, &friend.FriendInfo{
			Id:        int64(f.Id),
			FriendId:  int64(f.FriendId),
			Remark:    f.Remark,
			Status:    f.Status,
			CreatedAt: f.CreatedAt.Unix(),
		})
	}

	return &friend.GetFriendListResp{
		List:  list,
		Total: total,
	}, nil
}
