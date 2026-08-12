package post

import (
	"time"

	"github.com/google/uuid"

	postDomain "github.com/pablo/backend/internal/domain/post"
)

type PostDTO struct {
	ID          uuid.UUID  `db:"id"`
	Kind        string     `db:"kind"`
	Content     string     `db:"content"`
	Status      string     `db:"status"`
	ScheduledAt *time.Time `db:"scheduled_at"`
	PublishedAt *time.Time `db:"published_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

func NewPostFromDTO(dto PostDTO) postDomain.Post {
	return postDomain.NewPostWithID(
		postDomain.PostID(dto.ID),
		postDomain.Kind(dto.Kind),
		dto.Content,
		postDomain.Status(dto.Status),
		dto.ScheduledAt,
		dto.PublishedAt,
		dto.CreatedAt,
		dto.UpdatedAt,
	)
}

func NewPostsListFromDTO(dtos []PostDTO) []postDomain.Post {
	res := make([]postDomain.Post, 0, len(dtos))
	for _, dto := range dtos {
		res = append(res, NewPostFromDTO(dto))
	}
	return res
}
