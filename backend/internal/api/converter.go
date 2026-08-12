package api

import (
	accountDomain "github.com/pablo/backend/internal/domain/account"
	mediaDomain "github.com/pablo/backend/internal/domain/media"
	postDomain "github.com/pablo/backend/internal/domain/post"
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
