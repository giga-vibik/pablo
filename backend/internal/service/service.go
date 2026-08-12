package service

import (
	"github.com/pablo/backend/internal/config"
	"github.com/pablo/backend/internal/integration"
	"github.com/pablo/backend/internal/service/account"
	"github.com/pablo/backend/internal/service/media"
	"github.com/pablo/backend/internal/service/post"
	"github.com/pablo/backend/internal/storage"
)

type Services struct {
	PostService    post.PostService
	MediaService   media.MediaService
	AccountService account.AccountService
}

func NewServiceRegistry(
	cfg *config.Config,
	storageRegistry *storage.Storages,
	integrationRegistry *integration.Integrations,
) *Services {
	return &Services{
		PostService: post.NewPostService(
			storageRegistry.PostStorage,
			storageRegistry.PostTargetStorage,
			storageRegistry.MediaStorage,
			integrationRegistry.ZernioClient,
		),

		MediaService: media.NewMediaService(
			storageRegistry.MediaStorage,
			integrationRegistry.S3Storage,
		),

		AccountService: account.NewAccountService(
			storageRegistry.AccountStorage,
			integrationRegistry.ZernioClient,
		),
	}
}
