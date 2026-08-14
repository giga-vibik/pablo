package post

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	mediaDomain "github.com/pablo/backend/internal/domain/media"
	postDomain "github.com/pablo/backend/internal/domain/post"
	"github.com/pablo/backend/internal/integration/zernio"
)

var (
	ErrNoTargets    = errors.New("post must have at least one target platform")
	ErrNoVideo      = errors.New("video post must have a video attached")
	ErrTextPlatform = errors.New("threads accepts text only, video platforms require a video")
)

type PostStorage interface {
	CreatePost(ctx context.Context, post postDomain.Post) error
	GetPostByID(ctx context.Context, postID postDomain.PostID) (postDomain.Post, error)
	ListPosts(ctx context.Context, limit, offset int) ([]postDomain.Post, error)
	UpdatePost(ctx context.Context, post postDomain.Post) error
	DeletePost(ctx context.Context, postID postDomain.PostID) error
	ListDuePosts(ctx context.Context, now time.Time, limit int) ([]postDomain.Post, error)
}

type PostTargetStorage interface {
	CreateTargets(ctx context.Context, targets []postDomain.Target) error
	GetTargetsByPostID(ctx context.Context, postID postDomain.PostID) ([]postDomain.Target, error)
	GetTargetsByPostIDs(ctx context.Context, postIDs []postDomain.PostID) (map[postDomain.PostID][]postDomain.Target, error)
	UpdateTarget(ctx context.Context, target postDomain.Target) error
}

type MediaStorage interface {
	GetMediaByPostID(ctx context.Context, postID postDomain.PostID) ([]mediaDomain.Media, error)
	GetMediaByPostIDs(ctx context.Context, postIDs []postDomain.PostID) (map[postDomain.PostID][]mediaDomain.Media, error)
}

// PostWithMedia — пост вместе с прикреплённым видео. Список постов рисуется
// превьюшками, поэтому медиа нужно сразу, а не отдельным запросом на каждый пост.
type PostWithMedia struct {
	Post  postDomain.Post
	Media []mediaDomain.Media
}

// ZernioClient — внешний агрегатор публикаций.
type ZernioClient interface {
	IsConfigured() bool
	ResolveAccountID(ctx context.Context, platform string) (string, error)
	Post(ctx context.Context, req zernio.PostRequest) (zernio.PostResult, error)
	GetPostAnalytics(ctx context.Context, postID string) (zernio.PostAnalytics, error)
}

type PostService interface {
	CreatePost(ctx context.Context, post postDomain.Post, targets []postDomain.Target) error
	GetPost(ctx context.Context, postID postDomain.PostID) (postDomain.Post, []mediaDomain.Media, error)
	ListPosts(ctx context.Context, limit, offset int) ([]PostWithMedia, error)
	DeletePost(ctx context.Context, postID postDomain.PostID) error
	// PublishPost публикует пост во все его площадки прямо сейчас.
	PublishPost(ctx context.Context, postID postDomain.PostID) (postDomain.Post, error)
	// ListDuePosts забирает посты, которым подошло время (для воркера).
	ListDuePosts(ctx context.Context, limit int) ([]postDomain.Post, error)
	// FailPost помечает пост упавшим.
	FailPost(ctx context.Context, postID postDomain.PostID) error
	// GetPostStats тянет статистику поста из zernio, см. stats.go.
	GetPostStats(ctx context.Context, postID postDomain.PostID) (PostStats, error)
}

type service struct {
	postStorage       PostStorage
	postTargetStorage PostTargetStorage
	mediaStorage      MediaStorage
	zernioClient      ZernioClient
}

func NewPostService(
	postStorage PostStorage,
	postTargetStorage PostTargetStorage,
	mediaStorage MediaStorage,
	zernioClient ZernioClient,
) PostService {
	return &service{
		postStorage:       postStorage,
		postTargetStorage: postTargetStorage,
		mediaStorage:      mediaStorage,
		zernioClient:      zernioClient,
	}
}

func (s *service) CreatePost(ctx context.Context, post postDomain.Post, targets []postDomain.Target) error {
	if len(targets) == 0 {
		return ErrNoTargets
	}

	// Видео-пост не может таргетиться в threads, текстовый — в видео-площадки.
	for _, t := range targets {
		if post.GetKind() == postDomain.VideoKind && !t.GetPlatform().IsVideo() {
			return ErrTextPlatform
		}
		if post.GetKind() == postDomain.TextKind && t.GetPlatform().IsVideo() {
			return ErrTextPlatform
		}
	}

	if err := s.postStorage.CreatePost(ctx, post); err != nil {
		log.Println("error: while creating post", err.Error())
		return err
	}

	if err := s.postTargetStorage.CreateTargets(ctx, targets); err != nil {
		log.Println("error: while creating post targets", err.Error())
		return err
	}

	return nil
}

func (s *service) GetPost(
	ctx context.Context,
	postID postDomain.PostID,
) (postDomain.Post, []mediaDomain.Media, error) {
	post, err := s.postStorage.GetPostByID(ctx, postID)
	if err != nil {
		return postDomain.Post{}, nil, err
	}

	targets, err := s.postTargetStorage.GetTargetsByPostID(ctx, postID)
	if err != nil {
		return postDomain.Post{}, nil, err
	}
	post.SetTargets(targets)

	media, err := s.mediaStorage.GetMediaByPostID(ctx, postID)
	if err != nil {
		return postDomain.Post{}, nil, err
	}

	return post, media, nil
}

func (s *service) ListPosts(ctx context.Context, limit, offset int) ([]PostWithMedia, error) {
	posts, err := s.postStorage.ListPosts(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return []PostWithMedia{}, nil
	}

	ids := make([]postDomain.PostID, 0, len(posts))
	for i := range posts {
		ids = append(ids, posts[i].GetID())
	}

	targetsByPost, err := s.postTargetStorage.GetTargetsByPostIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	mediaByPost, err := s.mediaStorage.GetMediaByPostIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	res := make([]PostWithMedia, 0, len(posts))
	for i := range posts {
		postID := posts[i].GetID()
		posts[i].SetTargets(targetsByPost[postID])

		res = append(res, PostWithMedia{
			Post:  posts[i],
			Media: mediaByPost[postID],
		})
	}

	return res, nil
}

func (s *service) DeletePost(ctx context.Context, postID postDomain.PostID) error {
	// Таргеты и медиа удалятся каскадом (ON DELETE CASCADE).
	return s.postStorage.DeletePost(ctx, postID)
}

func (s *service) ListDuePosts(ctx context.Context, limit int) ([]postDomain.Post, error) {
	posts, err := s.postStorage.ListDuePosts(ctx, time.Now().UTC(), limit)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		targets, tErr := s.postTargetStorage.GetTargetsByPostID(ctx, posts[i].GetID())
		if tErr != nil {
			log.Printf("error: while loading targets for post %s: %s", posts[i].GetID().String(), tErr.Error())
			continue
		}
		posts[i].SetTargets(targets)
	}

	return posts, nil
}

// FailPost нужен воркеру: ListDuePosts уже перевёл пост в publishing, и если
// публикация упала до отправки, пост так и завис бы в publishing — из выборки
// due он уже вышел, а значит никто бы его больше не тронул.
func (s *service) FailPost(ctx context.Context, postID postDomain.PostID) error {
	post, err := s.postStorage.GetPostByID(ctx, postID)
	if err != nil {
		return err
	}

	post.SetStatus(postDomain.FailedStatus)

	return s.postStorage.UpdatePost(ctx, post)
}

// PublishPost отправляет пост в каждую площадку отдельным вызовом zernio.
// Отдельно, а не одним запросом со списком платформ, чтобы падение одной
// площадки не отменяло остальные и статус вёлся по каждой независимо.
func (s *service) PublishPost(ctx context.Context, postID postDomain.PostID) (postDomain.Post, error) {
	post, media, err := s.GetPost(ctx, postID)
	if err != nil {
		return postDomain.Post{}, err
	}

	if !s.zernioClient.IsConfigured() {
		return postDomain.Post{}, zernio.ErrNotConfigured
	}

	targets := post.GetTargets()
	if len(targets) == 0 {
		return postDomain.Post{}, ErrNoTargets
	}

	// Видео-пост без файла публиковать нечем.
	var mediaItems []zernio.MediaItem
	if post.GetKind() == postDomain.VideoKind {
		if len(media) == 0 || media[0].GetPublicURL() == "" {
			return postDomain.Post{}, ErrNoVideo
		}
		mediaItems = []zernio.MediaItem{{Type: "video", URL: media[0].GetPublicURL()}}
	}

	post.SetStatus(postDomain.PublishingStatus)
	if err = s.postStorage.UpdatePost(ctx, post); err != nil {
		log.Println("error: while marking post as publishing", err.Error())
	}

	for i := range targets {
		s.publishTarget(ctx, &targets[i], post, mediaItems)

		if uErr := s.postTargetStorage.UpdateTarget(ctx, targets[i]); uErr != nil {
			log.Printf("error: while updating target %s: %s", targets[i].GetID().String(), uErr.Error())
		}
	}

	post.SetTargets(targets)
	post.SetStatus(post.ResolveStatus())

	if post.GetStatus() == postDomain.PublishedStatus || post.GetStatus() == postDomain.PartiallyPublishedStatus {
		now := time.Now().UTC()
		post.SetPublishedAt(&now)
	}

	if err = s.postStorage.UpdatePost(ctx, post); err != nil {
		log.Println("error: while updating post status", err.Error())
		return postDomain.Post{}, err
	}

	return post, nil
}

// publishTarget публикует один таргет и проставляет ему статус.
func (s *service) publishTarget(
	ctx context.Context,
	target *postDomain.Target,
	post postDomain.Post,
	mediaItems []zernio.MediaItem,
) {
	platform := string(target.GetPlatform())

	accountID, err := s.zernioClient.ResolveAccountID(ctx, platform)
	if err != nil {
		target.MarkFailed(fmt.Sprintf("resolve account: %s", err.Error()))
		return
	}

	zTarget := zernio.PostTarget{Platform: platform, AccountID: accountID}

	// Instagram: без contentType=reels видео уйдёт обычным постом в ленту.
	if target.GetPlatform() == postDomain.InstagramPlatform {
		zTarget.PlatformSpecificData = map[string]any{"contentType": "reels"}
	}

	// Текст берём из таргета — у каждой площадки свои лимиты и хештеги.
	// Если для площадки текст не задан, падаем на общий текст поста.
	content := target.GetCaption()
	if content == "" {
		content = post.GetContent()
	}

	res, err := s.zernioClient.Post(ctx, zernio.PostRequest{
		Content:    content,
		MediaItems: mediaItems,
		Platforms:  []zernio.PostTarget{zTarget},
		PublishNow: true,
		Timezone:   "UTC",
	})
	if err != nil {
		// 409 значит, что предыдущая попытка уже прошла — считаем успехом,
		// иначе ретрай будет вечно "падать" на уже опубликованном посте.
		if errors.Is(err, zernio.ErrDuplicateContent) {
			log.Printf("post %s target %s: duplicate content, treating as published",
				post.GetID().String(), platform)
			target.MarkPublished(target.GetExternalPostID(), target.GetExternalURL())
			return
		}
		target.MarkFailed(err.Error())
		return
	}

	// zernio отчитывается по каждому аккаунту отдельно.
	if len(res.Platforms) > 0 {
		pr := res.Platforms[0]

		if zernio.IsFailureStatus(pr.Status) {
			msg := pr.Error
			if msg == "" {
				msg = "zernio: post failed (status " + pr.Status + ")"
			}
			target.MarkFailed(msg)
			return
		}

		// Публикация асинхронная: publishing/processing — это принято, не ошибка.
		if zernio.IsPublishedStatus(pr.Status) {
			target.MarkPublished(res.ID, pr.PlatformPostURL)
		} else {
			target.MarkPublishing(res.ID)
		}

		return
	}

	if zernio.IsFailureStatus(res.Status) {
		target.MarkFailed("zernio: post failed (status " + res.Status + ")")
		return
	}

	if zernio.IsPublishedStatus(res.Status) {
		target.MarkPublished(res.ID, "")
	} else {
		target.MarkPublishing(res.ID)
	}
}
