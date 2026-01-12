package svc

import (
	"context"
	"log"

	"SkyeIM/app/upload/api/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ServiceContext struct {
	Config      config.Config
	MinIOClient *minio.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化MinIO客户端
	minioClient, err := minio.New(c.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.MinIO.AccessKeyID, c.MinIO.SecretAccessKey, ""),
		Secure: c.MinIO.UseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	// 创建buckets
	ctx := context.Background()
	buckets := []string{
		c.MinIO.Buckets.Avatar,
		c.MinIO.Buckets.Image,
		c.MinIO.Buckets.File,
		c.MinIO.Buckets.Video,
		c.MinIO.Buckets.Voice,
	}

	for _, bucket := range buckets {
		exists, err := minioClient.BucketExists(ctx, bucket)
		if err != nil {
			log.Fatalf("Failed to check bucket %s: %v", bucket, err)
		}
		if !exists {
			err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				log.Fatalf("Failed to create bucket %s: %v", bucket, err)
			}
			log.Printf("Created bucket: %s", bucket)
		}

		// 设置bucket为公开读
		policy := `{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::` + bucket + `/*"]
			}]
		}`
		err = minioClient.SetBucketPolicy(ctx, bucket, policy)
		if err != nil {
			log.Printf("Warning: Failed to set bucket policy for %s: %v", bucket, err)
		}
	}

	return &ServiceContext{
		Config:      c,
		MinIOClient: minioClient,
	}
}
