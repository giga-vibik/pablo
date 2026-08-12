package post_target

import (
	"time"

	"github.com/google/uuid"

	postDomain "github.com/pablo/backend/internal/domain/post"
)

type TargetDTO struct {
	ID             uuid.UUID  `db:"id"`
	PostID         uuid.UUID  `db:"post_id"`
	Platform       string     `db:"platform"`
	Caption        string     `db:"caption"`
	Status         string     `db:"status"`
	ExternalPostID string     `db:"external_post_id"`
	ExternalURL    string     `db:"external_url"`
	ErrorMessage   string     `db:"error_message"`
	PublishedAt    *time.Time `db:"published_at"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

func NewTargetFromDTO(dto TargetDTO) postDomain.Target {
	return postDomain.NewTargetWithID(
		postDomain.TargetID(dto.ID),
		postDomain.PostID(dto.PostID),
		postDomain.Platform(dto.Platform),
		dto.Caption,
		postDomain.TargetStatus(dto.Status),
		dto.ExternalPostID,
		dto.ExternalURL,
		dto.ErrorMessage,
		dto.PublishedAt,
		dto.CreatedAt,
		dto.UpdatedAt,
	)
}

func NewTargetsListFromDTO(dtos []TargetDTO) []postDomain.Target {
	res := make([]postDomain.Target, 0, len(dtos))
	for _, dto := range dtos {
		res = append(res, NewTargetFromDTO(dto))
	}
	return res
}
