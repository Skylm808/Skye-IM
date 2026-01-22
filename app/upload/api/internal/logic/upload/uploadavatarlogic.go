package upload

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"SkyeIM/app/upload/api/internal/svc"
	"SkyeIM/app/upload/api/internal/types"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/logx"
)

type UploadAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 上传头像（用户/群组）
func NewUploadAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadAvatarLogic {
	return &UploadAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadAvatarLogic) UploadAvatar(file *multipart.FileHeader) (resp *types.UploadAvatarResp, err error) {
	// 1. 验证文件类型
	contentType := file.Header.Get("Content-Type")
	if !IsAllowedImageType(contentType, l.svcCtx.Config.Upload.AllowedImageTypes) {
		return nil, fmt.Errorf("不支持的图片格式: %s", contentType)
	}

	// 2. 验证文件大小
	if file.Size > l.svcCtx.Config.Upload.MaxAvatarSize {
		return nil, fmt.Errorf("头像大小超过限制: %d MB", l.svcCtx.Config.Upload.MaxAvatarSize/1024/1024)
	}

	// 3. 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	objectName := fmt.Sprintf("avatars/%s", filename)

	// 4. 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 5. 上传到MinIO
	bucketName := l.svcCtx.Config.MinIO.Buckets.Avatar
	_, err = l.svcCtx.MinIOClient.PutObject(
		l.ctx,
		bucketName,
		objectName,
		src,
		file.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return nil, fmt.Errorf("上传失败: %w", err)
	}

	// 6. 生成URL（使用PublicEndpoint让前端可访问）
	publicEndpoint := l.svcCtx.Config.MinIO.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = l.svcCtx.Config.MinIO.Endpoint // 降级方案
	}
	url := fmt.Sprintf("http://%s/%s/%s", publicEndpoint, bucketName, objectName)

	// TODO: 生成缩略图（可选）
	thumbnail := url

	return &types.UploadAvatarResp{
		Url:       url,
		Thumbnail: thumbnail,
	}, nil
}

// IsAllowedImageType 检查图片类型是否允许
func IsAllowedImageType(contentType string, allowed []string) bool {
	for _, t := range allowed {
		if t == contentType {
			return true
		}
	}
	return false
}
