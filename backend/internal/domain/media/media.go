// Package media описывает видеофайл, прикреплённый к посту.
//
// Важно: zernio НЕ принимает файл — он забирает медиа по публичному URL.
// Поэтому видео сначала уезжает в S3, и в zernio уходит publicURL.
package media

import (
	"time"

	"github.com/google/uuid"

	postDomain "github.com/pablo/backend/internal/domain/post"
)

type MediaID uuid.UUID

func (m MediaID) String() string { return uuid.UUID(m).String() }

type Media struct {
	id         MediaID
	postID     postDomain.PostID
	fileName   string
	storageKey string
	publicURL  string
	mimeType   string
	sizeBytes  int64
	createdAt  time.Time

	content []byte // не хранится в БД, только для загрузки в S3
}

func NewMedia(postID postDomain.PostID, fileName, mimeType string, sizeBytes int64, content []byte) Media {
	id := MediaID(uuid.New())

	return Media{
		id:         id,
		postID:     postID,
		fileName:   fileName,
		storageKey: id.String() + extensionFor(mimeType),
		mimeType:   mimeType,
		sizeBytes:  sizeBytes,
		createdAt:  time.Now().UTC(),
		content:    content,
	}
}

func NewMediaWithID(
	id MediaID,
	postID postDomain.PostID,
	fileName string,
	storageKey string,
	publicURL string,
	mimeType string,
	sizeBytes int64,
	createdAt time.Time,
) Media {
	return Media{
		id:         id,
		postID:     postID,
		fileName:   fileName,
		storageKey: storageKey,
		publicURL:  publicURL,
		mimeType:   mimeType,
		sizeBytes:  sizeBytes,
		createdAt:  createdAt,
	}
}

func (m *Media) GetID() MediaID               { return m.id }
func (m *Media) GetPostID() postDomain.PostID { return m.postID }
func (m *Media) GetFileName() string          { return m.fileName }
func (m *Media) GetStorageKey() string        { return m.storageKey }
func (m *Media) GetPublicURL() string         { return m.publicURL }
func (m *Media) GetMimeType() string          { return m.mimeType }
func (m *Media) GetSizeBytes() int64          { return m.sizeBytes }
func (m *Media) GetCreatedAt() time.Time      { return m.createdAt }
func (m *Media) GetContent() []byte           { return m.content }

func (m *Media) SetPublicURL(u string) { m.publicURL = u }
func (m *Media) SetContent(c []byte)   { m.content = c }

func extensionFor(mimeType string) string {
	switch mimeType {
	case "video/quicktime":
		return ".mov"
	default:
		return ".mp4"
	}
}
