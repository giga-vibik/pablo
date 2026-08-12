package media

import (
	"context"
	"errors"
	"log"

	mediaDomain "github.com/pablo/backend/internal/domain/media"
	postDomain "github.com/pablo/backend/internal/domain/post"
)

var ErrEmptyFile = errors.New("empty file")

type MediaStorage interface {
	CreateMedia(ctx context.Context, m mediaDomain.Media) error
	GetMediaByPostID(ctx context.Context, postID postDomain.PostID) ([]mediaDomain.Media, error)
	DeleteMediaByPostID(ctx context.Context, postID postDomain.PostID) error
}

type S3Storage interface {
	SaveVideo(ctx context.Context, m mediaDomain.Media) (string, error)
	DeleteObject(ctx context.Context, storageKey string) error
}

type MediaService interface {
	// UploadVideo кладёт файл в S3 и сохраняет ссылку. Возвращает публичный URL,
	// по которому zernio заберёт видео при публикации.
	UploadVideo(
		ctx context.Context,
		postID postDomain.PostID,
		fileName string,
		mimeType string,
		content []byte,
	) (mediaDomain.Media, error)
	GetPostMedia(ctx context.Context, postID postDomain.PostID) ([]mediaDomain.Media, error)
}

type service struct {
	mediaStorage MediaStorage
	s3Storage    S3Storage
}

func NewMediaService(mediaStorage MediaStorage, s3Storage S3Storage) MediaService {
	return &service{
		mediaStorage: mediaStorage,
		s3Storage:    s3Storage,
	}
}

func (s *service) UploadVideo(
	ctx context.Context,
	postID postDomain.PostID,
	fileName string,
	mimeType string,
	content []byte,
) (mediaDomain.Media, error) {
	if len(content) == 0 {
		return mediaDomain.Media{}, ErrEmptyFile
	}

	// К посту прикрепляется одно видео — старое заменяем.
	old, err := s.mediaStorage.GetMediaByPostID(ctx, postID)
	if err != nil {
		return mediaDomain.Media{}, err
	}

	m := mediaDomain.NewMedia(postID, fileName, mimeType, int64(len(content)), content)

	publicURL, err := s.s3Storage.SaveVideo(ctx, m)
	if err != nil {
		log.Println("error: while uploading video to s3", err.Error())
		return mediaDomain.Media{}, err
	}
	m.SetPublicURL(publicURL)

	if err = s.mediaStorage.DeleteMediaByPostID(ctx, postID); err != nil {
		return mediaDomain.Media{}, err
	}

	if err = s.mediaStorage.CreateMedia(ctx, m); err != nil {
		log.Println("error: while saving media", err.Error())
		return mediaDomain.Media{}, err
	}

	// Чистим бакет только после успешной записи: если удалить раньше и запись
	// упадёт, у поста останется ссылка на несуществующий объект.
	for i := range old {
		if delErr := s.s3Storage.DeleteObject(ctx, old[i].GetStorageKey()); delErr != nil {
			log.Println("error: while deleting old media from s3", delErr.Error())
		}
	}

	return m, nil
}

func (s *service) GetPostMedia(ctx context.Context, postID postDomain.PostID) ([]mediaDomain.Media, error) {
	return s.mediaStorage.GetMediaByPostID(ctx, postID)
}
