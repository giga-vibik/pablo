// Package account — подключённый аккаунт соцсети.
//
// Источник истины — zernio: аккаунты подключаются через его hosted-OAuth.
// Локальная таблица нужна как кэш, чтобы фронт не ходил во внешний API на
// каждый рендер.
package account

import (
	"time"

	"github.com/google/uuid"

	postDomain "github.com/pablo/backend/internal/domain/post"
)

type AccountID uuid.UUID

func (a AccountID) String() string { return uuid.UUID(a).String() }

type Account struct {
	id              AccountID
	platform        postDomain.Platform
	zernioAccountID string
	username        string
	isActive        bool
	syncedAt        time.Time
	createdAt       time.Time
}

func NewAccount(platform postDomain.Platform, zernioAccountID, username string, isActive bool) Account {
	now := time.Now().UTC()

	return Account{
		id:              AccountID(uuid.New()),
		platform:        platform,
		zernioAccountID: zernioAccountID,
		username:        username,
		isActive:        isActive,
		syncedAt:        now,
		createdAt:       now,
	}
}

func NewAccountWithID(
	id AccountID,
	platform postDomain.Platform,
	zernioAccountID string,
	username string,
	isActive bool,
	syncedAt time.Time,
	createdAt time.Time,
) Account {
	return Account{
		id:              id,
		platform:        platform,
		zernioAccountID: zernioAccountID,
		username:        username,
		isActive:        isActive,
		syncedAt:        syncedAt,
		createdAt:       createdAt,
	}
}

func (a *Account) GetID() AccountID                 { return a.id }
func (a *Account) GetPlatform() postDomain.Platform { return a.platform }
func (a *Account) GetZernioAccountID() string       { return a.zernioAccountID }
func (a *Account) GetUsername() string              { return a.username }
func (a *Account) IsActive() bool                   { return a.isActive }
func (a *Account) GetSyncedAt() time.Time           { return a.syncedAt }
func (a *Account) GetCreatedAt() time.Time          { return a.createdAt }
