package media

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	mediaDomain "github.com/pablo/backend/internal/domain/media"
	postDomain "github.com/pablo/backend/internal/domain/post"
)

const mediaTableName = "media"

type Storage interface {
	CreateMedia(ctx context.Context, m mediaDomain.Media) error
	GetMediaByPostID(ctx context.Context, postID postDomain.PostID) ([]mediaDomain.Media, error)
	// GetMediaByPostIDs — батч для списка постов, чтобы не ходить в базу на каждый.
	GetMediaByPostIDs(ctx context.Context, postIDs []postDomain.PostID) (map[postDomain.PostID][]mediaDomain.Media, error)
	DeleteMediaByPostID(ctx context.Context, postID postDomain.PostID) error
}

type storage struct {
	db *sqlx.DB
}

func NewMediaStorage(db *sqlx.DB) Storage {
	return &storage{db: db}
}

type MediaDTO struct {
	ID         uuid.UUID `db:"id"`
	PostID     uuid.UUID `db:"post_id"`
	FileName   string    `db:"file_name"`
	StorageKey string    `db:"storage_key"`
	PublicURL  string    `db:"public_url"`
	MimeType   string    `db:"mime_type"`
	SizeBytes  int64     `db:"size_bytes"`
	CreatedAt  time.Time `db:"created_at"`
}

func newMediaFromDTO(dto MediaDTO) mediaDomain.Media {
	return mediaDomain.NewMediaWithID(
		mediaDomain.MediaID(dto.ID),
		postDomain.PostID(dto.PostID),
		dto.FileName,
		dto.StorageKey,
		dto.PublicURL,
		dto.MimeType,
		dto.SizeBytes,
		dto.CreatedAt,
	)
}

func (s *storage) CreateMedia(ctx context.Context, m mediaDomain.Media) error {
	query := squirrel.Insert(mediaTableName).
		Columns("id", "post_id", "file_name", "storage_key", "public_url", "mime_type", "size_bytes", "created_at").
		Values(
			m.GetID().String(),
			m.GetPostID().String(),
			m.GetFileName(),
			m.GetStorageKey(),
			m.GetPublicURL(),
			m.GetMimeType(),
			m.GetSizeBytes(),
			m.GetCreatedAt(),
		).
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlQuery, args...)
	return err
}

func (s *storage) GetMediaByPostID(ctx context.Context, postID postDomain.PostID) ([]mediaDomain.Media, error) {
	query := squirrel.Select("id", "post_id", "file_name", "storage_key", "public_url", "mime_type", "size_bytes", "created_at").
		From(mediaTableName).
		Where(squirrel.Eq{"post_id": postID.String()}).
		OrderBy("created_at").
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var dtos []MediaDTO
	if err = s.db.SelectContext(ctx, &dtos, sqlQuery, args...); err != nil {
		return nil, err
	}

	res := make([]mediaDomain.Media, 0, len(dtos))
	for _, dto := range dtos {
		res = append(res, newMediaFromDTO(dto))
	}

	return res, nil
}

func (s *storage) GetMediaByPostIDs(
	ctx context.Context,
	postIDs []postDomain.PostID,
) (map[postDomain.PostID][]mediaDomain.Media, error) {
	res := make(map[postDomain.PostID][]mediaDomain.Media, len(postIDs))
	if len(postIDs) == 0 {
		return res, nil
	}

	rawIDs := make([]string, 0, len(postIDs))
	for _, id := range postIDs {
		rawIDs = append(rawIDs, id.String())
	}

	query := squirrel.Select("id", "post_id", "file_name", "storage_key", "public_url", "mime_type", "size_bytes", "created_at").
		From(mediaTableName).
		Where(squirrel.Eq{"post_id": rawIDs}).
		OrderBy("created_at").
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var dtos []MediaDTO
	if err = s.db.SelectContext(ctx, &dtos, sqlQuery, args...); err != nil {
		return nil, err
	}

	for _, dto := range dtos {
		postID := postDomain.PostID(dto.PostID)
		res[postID] = append(res[postID], newMediaFromDTO(dto))
	}

	return res, nil
}

func (s *storage) DeleteMediaByPostID(ctx context.Context, postID postDomain.PostID) error {
	query := squirrel.Delete(mediaTableName).
		Where(squirrel.Eq{"post_id": postID.String()}).
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlQuery, args...)
	return err
}
