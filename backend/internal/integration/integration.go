package integration

import (
	"net/http"
	"time"

	"github.com/pablo/backend/internal/config"
	"github.com/pablo/backend/internal/integration/s3"
	"github.com/pablo/backend/internal/integration/zernio"
)

type Integrations struct {
	ZernioClient *zernio.Client
	S3Storage    s3.Storage
}

func NewIntegrationRegistry(cfg *config.Config) (*Integrations, error) {
	// Публикация видео идёт долго — короткий дефолтный таймаут её рвёт.
	httpClient := &http.Client{Timeout: 60 * time.Second}

	s3Storage, err := s3.NewS3Storage(cfg.S3)
	if err != nil {
		return nil, err
	}

	return &Integrations{
		ZernioClient: zernio.New(cfg.Zernio, httpClient),
		S3Storage:    s3Storage,
	}, nil
}
