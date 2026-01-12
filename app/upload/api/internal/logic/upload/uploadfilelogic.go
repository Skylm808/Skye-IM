package upload

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"SkyeIM/app/upload/api/internal/svc"
	"SkyeIM/app/upload/api/internal/types"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/logx"
)

type UploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 上传文件（文档/视频等）
func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadFileLogic {
	return &UploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadFileLogic) UploadFile(file *multipart.FileHeader) (resp *types.UploadFileResp, err error) {
	// 1. 验证文件大小
	if file.Size > l.svcCtx.Config.Upload.MaxFileSize {
		return nil, fmt.Errorf("文件大小超过限制: %d MB", l.svcCtx.Config.Upload.MaxFileSize/1024/1024)
	}

	// 2. 生成唯一文件名（保留原始文件名）
	ext := filepath.Ext(file.Filename)
	uuidStr := uuid.New().String()
	filename := fmt.Sprintf("%s%s", uuidStr, ext)
	objectName := fmt.Sprintf("%s/%s", time.Now().Format("2006/01/02"), filename)

	// 3. 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 4. 获取MIME类型
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 5. 上传到MinIO
	bucketName := l.svcCtx.Config.MinIO.Buckets.File
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

	// 6. 生成URL
	url := fmt.Sprintf("http://%s/%s/%s", l.svcCtx.Config.MinIO.Endpoint, bucketName, objectName)

	return &types.UploadFileResp{
		Url:      url,
		Filename: file.Filename,
		Size:     file.Size,
		MimeType: contentType,
	}, nil
}
