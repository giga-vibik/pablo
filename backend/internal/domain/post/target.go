package post

import (
	"time"

	"github.com/google/uuid"
)

type TargetID uuid.UUID

func (t TargetID) String() string { return uuid.UUID(t).String() }

// TargetStatus — статус публикации в конкретную площадку. Считается отдельно от
// поста: публикация в TikTok может пройти, а в Instagram упасть.
type TargetStatus string

const (
	TargetPendingStatus TargetStatus = "pending"
	// TargetPublishingStatus — zernio принял пост, но площадка ещё обрабатывает
	// его (Instagram публикует асинхронно). Не ошибка.
	TargetPublishingStatus TargetStatus = "publishing"
	TargetPublishedStatus  TargetStatus = "published"
	TargetFailedStatus     TargetStatus = "failed"
)

// Target — одна площадка внутри поста, со своим текстом и своим статусом.
type Target struct {
	id             TargetID
	postID         PostID
	platform       Platform
	caption        string
	status         TargetStatus
	externalPostID string
	externalURL    string
	errorMessage   string
	publishedAt    *time.Time
	createdAt      time.Time
	updatedAt      time.Time
}

func NewTarget(postID PostID, platform Platform, caption string) Target {
	now := time.Now().UTC()

	return Target{
		id:        TargetID(uuid.New()),
		postID:    postID,
		platform:  platform,
		caption:   caption,
		status:    TargetPendingStatus,
		createdAt: now,
		updatedAt: now,
	}
}

func NewTargetWithID(
	id TargetID,
	postID PostID,
	platform Platform,
	caption string,
	status TargetStatus,
	externalPostID string,
	externalURL string,
	errorMessage string,
	publishedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) Target {
	return Target{
		id:             id,
		postID:         postID,
		platform:       platform,
		caption:        caption,
		status:         status,
		externalPostID: externalPostID,
		externalURL:    externalURL,
		errorMessage:   errorMessage,
		publishedAt:    publishedAt,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (t *Target) GetID() TargetID            { return t.id }
func (t *Target) GetPostID() PostID          { return t.postID }
func (t *Target) GetPlatform() Platform      { return t.platform }
func (t *Target) GetCaption() string         { return t.caption }
func (t *Target) GetStatus() TargetStatus    { return t.status }
func (t *Target) GetExternalPostID() string  { return t.externalPostID }
func (t *Target) GetExternalURL() string     { return t.externalURL }
func (t *Target) GetErrorMessage() string    { return t.errorMessage }
func (t *Target) GetPublishedAt() *time.Time { return t.publishedAt }
func (t *Target) GetCreatedAt() time.Time    { return t.createdAt }

func (t *Target) SetCaption(c string)         { t.caption = c }
func (t *Target) SetStatus(s TargetStatus)    { t.status = s }
func (t *Target) SetExternalPostID(id string) { t.externalPostID = id }
func (t *Target) SetExternalURL(u string)     { t.externalURL = u }
func (t *Target) SetErrorMessage(m string)    { t.errorMessage = m }
func (t *Target) SetPublishedAt(v *time.Time) { t.publishedAt = v }

// MarkPublished фиксирует успешную отправку в площадку.
func (t *Target) MarkPublished(externalID, externalURL string) {
	now := time.Now().UTC()

	t.status = TargetPublishedStatus
	t.externalPostID = externalID
	t.externalURL = externalURL
	t.errorMessage = ""
	t.publishedAt = &now
}

// MarkPublishing — zernio принял пост, площадка ещё обрабатывает.
func (t *Target) MarkPublishing(externalID string) {
	t.status = TargetPublishingStatus
	t.externalPostID = externalID
	t.errorMessage = ""
}

func (t *Target) MarkFailed(msg string) {
	t.status = TargetFailedStatus
	t.errorMessage = msg
}
