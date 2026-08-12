package post_target

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	postDomain "github.com/pablo/backend/internal/domain/post"
)

const postTargetsTableName = "post_targets"

type Storage interface {
	CreateTargets(ctx context.Context, targets []postDomain.Target) error
	GetTargetsByPostID(ctx context.Context, postID postDomain.PostID) ([]postDomain.Target, error)
	GetTargetsByPostIDs(ctx context.Context, postIDs []postDomain.PostID) (map[postDomain.PostID][]postDomain.Target, error)
	UpdateTarget(ctx context.Context, target postDomain.Target) error
}

type storage struct {
	db *sqlx.DB
}

func NewPostTargetStorage(db *sqlx.DB) Storage {
	return &storage{db: db}
}

var targetColumns = []string{
	"id", "post_id", "platform", "caption", "status",
	"external_post_id", "external_url", "error_message",
	"published_at", "created_at", "updated_at",
}

func (s *storage) CreateTargets(ctx context.Context, targets []postDomain.Target) error {
	if len(targets) == 0 {
		return nil
	}

	query := squirrel.Insert(postTargetsTableName).Columns(targetColumns...)

	for _, t := range targets {
		query = query.Values(
			t.GetID().String(),
			t.GetPostID().String(),
			string(t.GetPlatform()),
			t.GetCaption(),
			string(t.GetStatus()),
			t.GetExternalPostID(),
			t.GetExternalURL(),
			t.GetErrorMessage(),
			t.GetPublishedAt(),
			t.GetCreatedAt(),
			t.GetCreatedAt(),
		)
	}

	sqlQuery, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlQuery, args...)
	return err
}

func (s *storage) GetTargetsByPostID(ctx context.Context, postID postDomain.PostID) ([]postDomain.Target, error) {
	query := squirrel.Select(targetColumns...).
		From(postTargetsTableName).
		Where(squirrel.Eq{"post_id": postID.String()}).
		OrderBy("platform").
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var dtos []TargetDTO
	if err = s.db.SelectContext(ctx, &dtos, sqlQuery, args...); err != nil {
		return nil, err
	}

	return NewTargetsListFromDTO(dtos), nil
}

// GetTargetsByPostIDs — батч для списка постов, чтобы не делать N+1 запросов.
func (s *storage) GetTargetsByPostIDs(
	ctx context.Context,
	postIDs []postDomain.PostID,
) (map[postDomain.PostID][]postDomain.Target, error) {
	res := make(map[postDomain.PostID][]postDomain.Target, len(postIDs))
	if len(postIDs) == 0 {
		return res, nil
	}

	ids := make([]string, 0, len(postIDs))
	for _, id := range postIDs {
		ids = append(ids, id.String())
	}

	query := squirrel.Select(targetColumns...).
		From(postTargetsTableName).
		Where(squirrel.Eq{"post_id": ids}).
		OrderBy("platform").
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var dtos []TargetDTO
	if err = s.db.SelectContext(ctx, &dtos, sqlQuery, args...); err != nil {
		return nil, err
	}

	for _, t := range NewTargetsListFromDTO(dtos) {
		res[t.GetPostID()] = append(res[t.GetPostID()], t)
	}

	return res, nil
}

func (s *storage) UpdateTarget(ctx context.Context, target postDomain.Target) error {
	query := squirrel.Update(postTargetsTableName).
		Set("caption", target.GetCaption()).
		Set("status", string(target.GetStatus())).
		Set("external_post_id", target.GetExternalPostID()).
		Set("external_url", target.GetExternalURL()).
		Set("error_message", target.GetErrorMessage()).
		Set("published_at", target.GetPublishedAt()).
		Set("updated_at", time.Now().UTC()).
		Where(squirrel.Eq{"id": uuid.UUID(target.GetID()).String()}).
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlQuery, args...)
	return err
}
