package post

import (
	"context"
	"errors"
	"log"
	"sort"
	"sync"

	postDomain "github.com/pablo/backend/internal/domain/post"
	"github.com/pablo/backend/internal/integration/zernio"
)

// StatsState — состояние статистики по площадке. Отдельно от цифр, потому что
// «нули» и «данных ещё нет» — принципиально разные вещи: первое можно показывать,
// второе нельзя, иначе пост выглядит провальным через минуту после публикации.
type StatsState string

const (
	// StatsReady — цифры пришли.
	StatsReady StatsState = "ready"
	// StatsPending — zernio ещё тянет данные с площадки.
	StatsPending StatsState = "pending"
	// StatsUnavailable — публикация не удалась либо статистика недоступна.
	StatsUnavailable StatsState = "unavailable"
)

type Metrics struct {
	Impressions    float64
	Reach          float64
	Views          float64
	Likes          float64
	Comments       float64
	Shares         float64
	Saves          float64
	Clicks         float64
	EngagementRate float64
}

type PlatformStats struct {
	Platform    postDomain.Platform
	State       StatsState
	Message     string
	Username    string
	URL         string
	LastUpdated string
	Metrics     Metrics
}

type PostStats struct {
	PostID    postDomain.PostID
	Totals    Metrics
	Platforms []PlatformStats
}

// GetPostStats собирает статистику поста по всем его площадкам.
//
// Один таргет = один пост в zernio (публикуем площадки по отдельности), поэтому
// и запросов столько же. Идут параллельно: последовательно четыре похода во
// внешний API складываются в заметную паузу перед поп-апом.
func (s *service) GetPostStats(ctx context.Context, postID postDomain.PostID) (PostStats, error) {
	post, err := s.postStorage.GetPostByID(ctx, postID)
	if err != nil {
		return PostStats{}, err
	}

	targets, err := s.postTargetStorage.GetTargetsByPostID(ctx, postID)
	if err != nil {
		return PostStats{}, err
	}

	if !s.zernioClient.IsConfigured() {
		return PostStats{}, zernio.ErrNotConfigured
	}

	res := PostStats{
		PostID:    post.GetID(),
		Platforms: make([]PlatformStats, len(targets)),
	}

	var wg sync.WaitGroup
	for i := range targets {
		wg.Add(1)

		go func(idx int, target postDomain.Target) {
			defer wg.Done()
			res.Platforms[idx] = s.targetStats(ctx, target)
		}(i, targets[i])
	}
	wg.Wait()

	// Суммируем только то, что реально пришло: складывать pending-нули значит
	// показывать заниженный итог как достоверный.
	for _, p := range res.Platforms {
		if p.State != StatsReady {
			continue
		}

		res.Totals.Impressions += p.Metrics.Impressions
		res.Totals.Reach += p.Metrics.Reach
		res.Totals.Views += p.Metrics.Views
		res.Totals.Likes += p.Metrics.Likes
		res.Totals.Comments += p.Metrics.Comments
		res.Totals.Shares += p.Metrics.Shares
		res.Totals.Saves += p.Metrics.Saves
		res.Totals.Clicks += p.Metrics.Clicks
	}

	// Вовлечённость — это доля, её нельзя складывать: считаем от суммарных цифр.
	if res.Totals.Impressions > 0 {
		engaged := res.Totals.Likes + res.Totals.Comments + res.Totals.Shares + res.Totals.Saves
		res.Totals.EngagementRate = engaged / res.Totals.Impressions * 100
	}

	sort.Slice(res.Platforms, func(i, j int) bool {
		return res.Platforms[i].Platform < res.Platforms[j].Platform
	})

	return res, nil
}

func (s *service) targetStats(ctx context.Context, target postDomain.Target) PlatformStats {
	stats := PlatformStats{
		Platform: target.GetPlatform(),
		URL:      target.GetExternalURL(),
		State:    StatsUnavailable,
	}

	if target.GetStatus() == postDomain.TargetFailedStatus {
		stats.Message = target.GetErrorMessage()
		if stats.Message == "" {
			stats.Message = "публикация не удалась"
		}

		return stats
	}

	externalID := target.GetExternalPostID()
	if externalID == "" {
		stats.Message = "пост ещё не отправлен в площадку"
		return stats
	}

	analytics, err := s.zernioClient.GetPostAnalytics(ctx, externalID)
	if err != nil {
		if errors.Is(err, zernio.ErrAnalyticsUnavailable) {
			stats.Message = "площадка не приняла публикацию"
			return stats
		}

		log.Printf("error: while getting analytics for %s: %s", externalID, err.Error())
		stats.Message = "не удалось получить статистику"

		return stats
	}

	// Ищем свою площадку в разбивке; если её нет — берём агрегат поста,
	// он для одноплощадочного поста zernio ему и равен.
	metrics := analytics.Analytics
	for _, p := range analytics.Platforms {
		if !equalPlatform(p.Platform, target.GetPlatform()) {
			continue
		}

		if p.Analytics != nil {
			metrics = p.Analytics
		}
		if p.AccountUsername != "" {
			stats.Username = p.AccountUsername
		}
		if p.PlatformPostURL != nil && *p.PlatformPostURL != "" {
			stats.URL = *p.PlatformPostURL
		}

		break
	}

	if analytics.IsPending() || metrics == nil {
		stats.State = StatsPending
		stats.Message = "zernio ещё синхронизирует данные с площадкой"

		return stats
	}

	stats.State = StatsReady
	stats.Metrics = Metrics{
		Impressions:    metrics.Impressions,
		Reach:          metrics.Reach,
		Views:          metrics.Views,
		Likes:          metrics.Likes,
		Comments:       metrics.Comments,
		Shares:         metrics.Shares,
		Saves:          metrics.Saves,
		Clicks:         metrics.Clicks,
		EngagementRate: metrics.EngagementRate,
	}

	if metrics.LastUpdated != nil {
		stats.LastUpdated = *metrics.LastUpdated
	}

	return stats
}

func equalPlatform(raw string, platform postDomain.Platform) bool {
	return postDomain.Platform(raw) == platform
}
