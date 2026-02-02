package repository

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"

	s3config "taskS3/internal/config"
)

type S3Repository interface {
	UploadFile(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	DownloadFile(ctx context.Context, key string) (io.ReadCloser, error)
	ListFiles(ctx context.Context, prefix string) ([]string, error)
	CopyFile(ctx context.Context, sourceKey, destKey string) error
}

type s3Repository struct {
	client *s3.Client
	cfg    *s3config.S3Config
	log    *zap.Logger
}

func NewS3Repository(cfg *s3config.S3Config, log *zap.Logger) (S3Repository, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.Endpoint != "" {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				HostnameImmutable: true,
				Source:            aws.EndpointSourceCustom,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	repo := &s3Repository{
		client: client,
		cfg:    cfg,
		log:    log,
	}

	// Проверяем и создаем bucket если нужно
	if err := repo.ensureBucketExists(context.Background()); err != nil {
		log.Warn("Failed to ensure bucket exists", zap.Error(err))
	}

	return repo, nil
}

func (r *s3Repository) ensureBucketExists(ctx context.Context) error {
	// Проверяем существование bucket
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.cfg.BucketName),
	})

	if err == nil {
		r.log.Info("Bucket already exists", zap.String("bucket", r.cfg.BucketName))
		return nil
	}

	// Создаем bucket
	r.log.Info("Creating bucket", zap.String("bucket", r.cfg.BucketName))

	_, err = r.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(r.cfg.BucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(r.cfg.Region),
		},
	})

	if err != nil {
		return err
	}

	r.log.Info("Bucket created successfully", zap.String("bucket", r.cfg.BucketName))

	// Даем время на создание
	time.Sleep(1 * time.Second)

	return nil
}

func (r *s3Repository) UploadFile(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.cfg.BucketName),
		Key:           aws.String(key),
		Body:          body,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})

	if err != nil {
		r.log.Error("Failed to upload file to S3",
			zap.String("key", key),
			zap.Error(err))
		return err
	}

	r.log.Info("File uploaded to S3",
		zap.String("key", key),
		zap.Int64("size", size))

	return nil
}

func (r *s3Repository) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.cfg.BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		r.log.Error("Failed to download file from S3",
			zap.String("key", key),
			zap.Error(err))
		return nil, err
	}

	return output.Body, nil
}

func (r *s3Repository) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	output, err := r.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(r.cfg.BucketName),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		return nil, err
	}

	var keys []string
	for _, obj := range output.Contents {
		keys = append(keys, *obj.Key)
	}

	return keys, nil
}

func (r *s3Repository) CopyFile(ctx context.Context, sourceKey, destKey string) error {
	_, err := r.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(r.cfg.BucketName),
		CopySource: aws.String(r.cfg.BucketName + "/" + sourceKey),
		Key:        aws.String(destKey),
	})

	if err != nil {
		r.log.Error("Failed to copy file in S3",
			zap.String("source", sourceKey),
			zap.String("destination", destKey),
			zap.Error(err))
		return err
	}

	r.log.Info("File copied in S3",
		zap.String("source", sourceKey),
		zap.String("destination", destKey))

	return nil
}
