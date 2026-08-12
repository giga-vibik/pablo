package post

import (
	"time"

	"github.com/google/uuid"
)

type PostID uuid.UUID

func (p PostID) String() string { return uuid.UUID(p).String() }

// Platform — площадка публикации. Значения совпадают с тем, что ждёт zernio.
type Platform string

const (
	InstagramPlatform Platform = "instagram" // reels
	TiktokPlatform    Platform = "tiktok"
	YoutubePlatform   Platform = "youtube" // shorts
	ThreadsPlatform   Platform = "threads" // только текст
)

// IsVideo сообщает, требует ли площадка видео. Threads — единственная текстовая.
func (p Platform) IsVideo() bool { return p != ThreadsPlatform }

func (p Platform) IsValid() bool {
	switch p {
	case InstagramPlatform, TiktokPlatform, YoutubePlatform, ThreadsPlatform:
		return true
	}
	return false
}

// Kind — тип поста. Определяет, нужно ли к нему видео.
type Kind string

const (
	VideoKind Kind = "video" // reels / tiktok / shorts
	TextKind  Kind = "text"  // threads
)

// Status — общий статус поста. Складывается из статусов таргетов:
// все published → published, часть failed → partially_published, все failed → failed.
type Status string

const (
	DraftStatus              Status = "draft"
	ScheduledStatus          Status = "scheduled"
	PublishingStatus         Status = "publishing"
	PublishedStatus          Status = "published"
	PartiallyPublishedStatus Status = "partially_published"
	FailedStatus             Status = "failed"
)

type Post struct {
	id          PostID
	kind        Kind
	content     string
	status      Status
	scheduledAt *time.Time
	publishedAt *time.Time
	createdAt   time.Time
	updatedAt   time.Time
	targets     []Target
}

func NewPost(kind Kind, content string, scheduledAt *time.Time) Post {
	now := time.Now().UTC()

	status := DraftStatus
	if scheduledAt != nil {
		status = ScheduledStatus
	}

	return Post{
		id:          PostID(uuid.New()),
		kind:        kind,
		content:     content,
		status:      status,
		scheduledAt: scheduledAt,
		createdAt:   now,
		updatedAt:   now,
	}
}

func NewPostWithID(
	id PostID,
	kind Kind,
	content string,
	status Status,
	scheduledAt *time.Time,
	publishedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) Post {
	return Post{
		id:          id,
		kind:        kind,
		content:     content,
		status:      status,
		scheduledAt: scheduledAt,
		publishedAt: publishedAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (p *Post) GetID() PostID              { return p.id }
func (p *Post) GetKind() Kind              { return p.kind }
func (p *Post) GetContent() string         { return p.content }
func (p *Post) GetStatus() Status          { return p.status }
func (p *Post) GetScheduledAt() *time.Time { return p.scheduledAt }
func (p *Post) GetPublishedAt() *time.Time { return p.publishedAt }
func (p *Post) GetCreatedAt() time.Time    { return p.createdAt }
func (p *Post) GetUpdatedAt() time.Time    { return p.updatedAt }
func (p *Post) GetTargets() []Target       { return p.targets }

func (p *Post) SetStatus(s Status)          { p.status = s }
func (p *Post) SetContent(c string)         { p.content = c }
func (p *Post) SetScheduledAt(t *time.Time) { p.scheduledAt = t }
func (p *Post) SetPublishedAt(t *time.Time) { p.publishedAt = t }
func (p *Post) SetTargets(t []Target)       { p.targets = t }

// ResolveStatus пересчитывает общий статус поста по статусам его таргетов.
func (p *Post) ResolveStatus() Status {
	if len(p.targets) == 0 {
		return p.status
	}

	var published, failed, pending int
	for _, t := range p.targets {
		switch t.GetStatus() {
		case TargetPublishedStatus:
			published++
		case TargetFailedStatus:
			failed++
		default:
			pending++
		}
	}

	switch {
	case pending > 0:
		return PublishingStatus
	case failed == len(p.targets):
		return FailedStatus
	case failed > 0:
		return PartiallyPublishedStatus
	default:
		return PublishedStatus
	}
}
