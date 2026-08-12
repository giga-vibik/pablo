package post

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	postDomain "github.com/pablo/backend/internal/domain/post"
)

const postsTableName = "posts"

type Storage interface {
	CreatePost(ctx context.Context, post postDomain.Post) error
	GetPostByID(ctx context.Context, postID postDomain.PostID) (postDomain.Post, error)
	ListPosts(ctx context.Context, limit, offset int) ([]postDomain.Post, error)
	UpdatePost(ctx context.Context, post postDomain.Post) error
	DeletePost(ctx context.Context, postID postDomain.PostID) error
	// ListDuePosts возвращает запланированные посты, которым подошло время.
	// SKIP LOCKED — чтобы несколько воркеров не взяли один пост.
	ListDuePosts(ctx context.Context, now time.Time, limit int) ([]postDomain.Post, error)
}

type storage struct {
	db *sqlx.DB
}

func NewPostStorage(db *sqlx.DB) Storage {
	return &storage{db: db}
}

var postColumns = []string{
	"id", "kind", "content", "status",
	"scheduled_at", "published_at", "created_at", "updated_at",
}

func (s *storage) CreatePost(ctx context.Context, post postDomain.Post) error {
	query := squirrel.Insert(postsTableName).
		Columns(postColumns...).
		Values(
			post.GetID().String(),
			string(post.GetKind()),
			post.GetContent(),
			string(post.GetStatus()),
			post.GetScheduledAt(),
			post.GetPublishedAt(),
			post.GetCreatedAt(),
			post.GetUpdatedAt(),
		).
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlQuery, args...)
	return err
}

func (s *storage) GetPostByID(ctx context.Context, postID postDomain.PostID) (postDomain.Post, error) {
	query := squirrel.Select(postColumns...).
		From(postsTableName).
		Where(squirrel.Eq{"id": postID.String()}).
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return postDomain.Post{}, err
	}

	var dto PostDTO
	if err = s.db.GetContext(ctx, &dto, sqlQuery, args...); err != nil {
		return postDomain.Post{}, err
	}

	return NewPostFromDTO(dto), nil
}

func (s *storage) ListPosts(ctx context.Context, limit, offset int) ([]postDomain.Post, error) {
	query := squirrel.Select(postColumns...).
		From(postsTableName).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var dtos []PostDTO
	if err = s.db.SelectContext(ctx, &dtos, sqlQuery, args...); err != nil {
		return nil, err
	}

	return NewPostsListFromDTO(dtos), nil
}

func (s *storage) UpdatePost(ctx context.Context, post postDomain.Post) error {
	query := squirrel.Update(postsTableName).
		Set("kind", string(post.GetKind())).
		Set("content", post.GetContent()).
		Set("status", string(post.GetStatus())).
		Set("scheduled_at", post.GetScheduledAt()).
		Set("published_at", post.GetPublishedAt()).
		Set("updated_at", time.Now().UTC()).
		Where(squirrel.Eq{"id": post.GetID().String()}).
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlQuery, args...)
	return err
}

func (s *storage) DeletePost(ctx context.Context, postID postDomain.PostID) error {
	query := squirrel.Delete(postsTableName).
		Where(squirrel.Eq{"id": postID.String()}).
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlQuery, args...)
	return err
}

// ListDuePosts забирает посты со статусом scheduled, у которых scheduled_at уже
// наступил, и сразу переводит их в publishing — так повторный тик воркера (или
// второй воркер) не возьмёт те же посты.
func (s *storage) ListDuePosts(ctx context.Context, now time.Time, limit int) ([]postDomain.Post, error) {
	const query = `
		UPDATE ` + postsTableName + ` AS p
		SET status = 'publishing', updated_at = $1
		WHERE p.id IN (
			SELECT id FROM ` + postsTableName + `
			WHERE status = 'scheduled' AND scheduled_at <= $1
			ORDER BY scheduled_at
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		RETURNING p.id, p.kind, p.content, p.status,
		          p.scheduled_at, p.published_at, p.created_at, p.updated_at`

	var dtos []PostDTO
	if err := s.db.SelectContext(ctx, &dtos, query, now, limit); err != nil {
		return nil, err
	}

	return NewPostsListFromDTO(dtos), nil
}
