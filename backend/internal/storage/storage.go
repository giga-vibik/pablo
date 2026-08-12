package storage

import (
	"github.com/jmoiron/sqlx"

	"github.com/pablo/backend/internal/storage/account"
	"github.com/pablo/backend/internal/storage/media"
	"github.com/pablo/backend/internal/storage/post"
	"github.com/pablo/backend/internal/storage/post_target"
)

type Storages struct {
	PostStorage       post.Storage
	PostTargetStorage post_target.Storage
	MediaStorage      media.Storage
	AccountStorage    account.Storage
}

func NewStorageRegistry(db *sqlx.DB) *Storages {
	return &Storages{
		PostStorage:       post.NewPostStorage(db),
		PostTargetStorage: post_target.NewPostTargetStorage(db),
		MediaStorage:      media.NewMediaStorage(db),
		AccountStorage:    account.NewAccountStorage(db),
	}
}
