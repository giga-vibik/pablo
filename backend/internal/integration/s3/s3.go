// Package s3 — загрузка видео в объектное хранилище.
//
// Объекты кладутся с публичным доступом: zernio скачивает медиа по URL, файл
// ему передать нельзя. Если бакет закрыт, публикация упадёт на стороне zernio.
package s3

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"

	"github.com/pablo/backend/internal/config"
	mediaDomain "github.com/pablo/backend/internal/domain/media"
)

type Storage interface {
	// SaveVideo кладёт файл в бакет и возвращает публичный URL.
	SaveVideo(ctx context.Context, m mediaDomain.Media) (string, error)
	DeleteObject(ctx context.Context, storageKey string) error
}

type storage struct {
	client     *awsS3.S3
	bucket     string
	publicBase string
}

func NewS3Storage(cfg config.S3) (Storage, error) {
	sess, err := session.NewSession(&awsSDK.Config{
		Region:      awsSDK.String(cfg.Region),
		Endpoint:    awsSDK.String(cfg.Endpoint),
		Credentials: credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("s3: new session: %w", err)
	}

	// Публичный адрес собираем в path-style (endpoint/bucket/key): так работает
	// и Yandex Object Storage, и MinIO, и любой другой S3-совместимый эндпоинт.
	publicBase := strings.TrimRight(cfg.Endpoint, "/") + "/" + cfg.BucketName

	return &storage{
		client:     awsS3.New(sess),
		bucket:     cfg.BucketName,
		publicBase: publicBase,
	}, nil
}

func (s *storage) SaveVideo(ctx context.Context, m mediaDomain.Media) (string, error) {
	_, err := s.client.PutObjectWithContext(ctx, &awsS3.PutObjectInput{
		Bucket:      awsSDK.String(s.bucket),
		Key:         awsSDK.String(m.GetStorageKey()),
		Body:        bytes.NewReader(m.GetContent()),
		ContentType: awsSDK.String(m.GetMimeType()),
		// zernio ходит за файлом анонимно — объект обязан быть публичным.
		ACL: awsSDK.String("public-read"),
	})
	if err != nil {
		return "", fmt.Errorf("s3: put object: %w", err)
	}

	return s.publicBase + "/" + m.GetStorageKey(), nil
}

func (s *storage) DeleteObject(ctx context.Context, storageKey string) error {
	_, err := s.client.DeleteObjectWithContext(ctx, &awsS3.DeleteObjectInput{
		Bucket: awsSDK.String(s.bucket),
		Key:    awsSDK.String(storageKey),
	})
	if err != nil {
		return fmt.Errorf("s3: delete object: %w", err)
	}

	return nil
}
