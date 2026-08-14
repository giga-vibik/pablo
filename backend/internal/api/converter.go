package api

import (
	accountDomain "github.com/pablo/backend/internal/domain/account"
	mediaDomain "github.com/pablo/backend/internal/domain/media"
	postDomain "github.com/pablo/backend/internal/domain/post"
	postService "github.com/pablo/backend/internal/service/post"
	"github.com/pablo/backend/schema"
)

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toSchemaPost(p postDomain.Post, media []mediaDomain.Media) schema.Post {
	targets := p.GetTargets()
	schemaTargets := make([]schema.Target, 0, len(targets))
	for i := range targets {
		schemaTargets = append(schemaTargets, toSchemaTarget(targets[i]))
	}

	res := schema.Post{
		Id:          p.GetID().String(),
		Kind:        schema.PostKind(p.GetKind()),
		Content:     optionalString(p.GetContent()),
		Status:      schema.PostStatus(p.GetStatus()),
		ScheduledAt: p.GetScheduledAt(),
		PublishedAt: p.GetPublishedAt(),
		Targets:     schemaTargets,
	}

	createdAt := p.GetCreatedAt()
	res.CreatedAt = &createdAt

	if len(media) > 0 {
		schemaMedia := make([]schema.Media, 0, len(media))
		for i := range media {
			schemaMedia = append(schemaMedia, toSchemaMedia(media[i]))
		}
		res.Media = &schemaMedia
	}

	return res
}

func toSchemaTarget(t postDomain.Target) schema.Target {
	return schema.Target{
		Id:             t.GetID().String(),
		Platform:       schema.Platform(t.GetPlatform()),
		Caption:        optionalString(t.GetCaption()),
		Status:         schema.TargetStatus(t.GetStatus()),
		ExternalPostId: optionalString(t.GetExternalPostID()),
		ExternalUrl:    optionalString(t.GetExternalURL()),
		ErrorMessage:   optionalString(t.GetErrorMessage()),
	}
}

func toSchemaMedia(m mediaDomain.Media) schema.Media {
	sizeBytes := m.GetSizeBytes()

	return schema.Media{
		Id:        m.GetID().String(),
		FileName:  m.GetFileName(),
		PublicUrl: m.GetPublicURL(),
		MimeType:  optionalString(m.GetMimeType()),
		SizeBytes: &sizeBytes,
	}
}

func toSchemaPostStats(stats postService.PostStats) schema.PostStats {
	platforms := make([]schema.PlatformStats, 0, len(stats.Platforms))
	for _, p := range stats.Platforms {
		metrics := toSchemaMetrics(p.Metrics)

		platforms = append(platforms, schema.PlatformStats{
			Platform:    schema.Platform(p.Platform),
			State:       schema.PlatformStatsState(p.State),
			Message:     optionalString(p.Message),
			Username:    optionalString(p.Username),
			Url:         optionalString(p.URL),
			LastUpdated: optionalString(p.LastUpdated),
			Metrics:     &metrics,
		})
	}

	totals := toSchemaMetrics(stats.Totals)

	return schema.PostStats{
		PostId:    stats.PostID.String(),
		Totals:    &totals,
		Platforms: platforms,
	}
}

func toSchemaMetrics(m postService.Metrics) schema.Metrics {
	f := func(v float64) *float32 {
		res := float32(v)
		return &res
	}

	return schema.Metrics{
		Impressions:    f(m.Impressions),
		Reach:          f(m.Reach),
		Views:          f(m.Views),
		Likes:          f(m.Likes),
		Comments:       f(m.Comments),
		Shares:         f(m.Shares),
		Saves:          f(m.Saves),
		Clicks:         f(m.Clicks),
		EngagementRate: f(m.EngagementRate),
	}
}

func toSchemaAccounts(accounts []accountDomain.Account) []schema.Account {
	res := make([]schema.Account, 0, len(accounts))
	for i := range accounts {
		res = append(res, toSchemaAccount(accounts[i]))
	}

	return res
}

func toSchemaAccount(a accountDomain.Account) schema.Account {
	syncedAt := a.GetSyncedAt()

	return schema.Account{
		Id:       a.GetID().String(),
		Platform: schema.Platform(a.GetPlatform()),
		Username: optionalString(a.GetUsername()),
		IsActive: a.IsActive(),
		SyncedAt: &syncedAt,
	}
}
