package logic

import (
	"context"
	"database/sql"

	"SkyeIM/app/user/rpc/internal/svc"
	"SkyeIM/app/user/rpc/user"
	"auth/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type FindUserByFieldLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFindUserByFieldLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindUserByFieldLogic {
	return &FindUserByFieldLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按字段查找用户（支持用户名/手机/邮箱）
func (l *FindUserByFieldLogic) FindUserByField(in *user.FindUserByFieldRequest) (*user.FindUserByFieldResponse, error) {
	var foundUser *model.User
	var err error

	switch in.FieldType {
	case "username":
		foundUser, err = l.svcCtx.UserModel.FindOneByUsername(l.ctx, in.FieldValue)
	case "phone":
		foundUser, err = l.svcCtx.UserModel.FindOneByPhone(l.ctx, sql.NullString{String: in.FieldValue, Valid: true})
	case "email":
		foundUser, err = l.svcCtx.UserModel.FindOneByEmail(l.ctx, sql.NullString{String: in.FieldValue, Valid: true})
	default:
		return &user.FindUserByFieldResponse{
			Found: false,
		}, nil
	}

	if err != nil {
		if err == model.ErrNotFound {
			return &user.FindUserByFieldResponse{
				Found: false,
			}, nil
		}
		l.Logger.Errorf("查询用户失败: fieldType=%s, fieldValue=%s, error=%v", in.FieldType, in.FieldValue, err)
		return nil, err
	}

	return &user.FindUserByFieldResponse{
		User:  convertToUserInfo(foundUser),
		Found: true,
	}, nil
}
