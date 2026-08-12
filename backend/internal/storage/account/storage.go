package account

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	accountDomain "github.com/pablo/backend/internal/domain/account"
	postDomain "github.com/pablo/backend/internal/domain/post"
)

const accountsTableName = "accounts"

type Storage interface {
	ListAccounts(ctx context.Context) ([]accountDomain.Account, error)
	// UpsertAccounts перезаписывает кэш аккаунтов данными из zernio.
	UpsertAccounts(ctx context.Context, accounts []accountDomain.Account) error
	// DeactivateMissing помечает неактивными аккаунты, которых больше нет в zernio.
	DeactivateMissing(ctx context.Context, presentZernioIDs []string) error
}

type storage struct {
	db *sqlx.DB
}

func NewAccountStorage(db *sqlx.DB) Storage {
	return &storage{db: db}
}

type AccountDTO struct {
	ID              uuid.UUID `db:"id"`
	Platform        string    `db:"platform"`
	ZernioAccountID string    `db:"zernio_account_id"`
	Username        string    `db:"username"`
	IsActive        bool      `db:"is_active"`
	SyncedAt        time.Time `db:"synced_at"`
	CreatedAt       time.Time `db:"created_at"`
}

func newAccountFromDTO(dto AccountDTO) accountDomain.Account {
	return accountDomain.NewAccountWithID(
		accountDomain.AccountID(dto.ID),
		postDomain.Platform(dto.Platform),
		dto.ZernioAccountID,
		dto.Username,
		dto.IsActive,
		dto.SyncedAt,
		dto.CreatedAt,
	)
}

func (s *storage) ListAccounts(ctx context.Context) ([]accountDomain.Account, error) {
	query := squirrel.Select("id", "platform", "zernio_account_id", "username", "is_active", "synced_at", "created_at").
		From(accountsTableName).
		OrderBy("platform").
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var dtos []AccountDTO
	if err = s.db.SelectContext(ctx, &dtos, sqlQuery, args...); err != nil {
		return nil, err
	}

	res := make([]accountDomain.Account, 0, len(dtos))
	for _, dto := range dtos {
		res = append(res, newAccountFromDTO(dto))
	}

	return res, nil
}

// UpsertAccounts вставляет новые и обновляет существующие по zernio_account_id.
func (s *storage) UpsertAccounts(ctx context.Context, accounts []accountDomain.Account) error {
	if len(accounts) == 0 {
		return nil
	}

	query := squirrel.Insert(accountsTableName).
		Columns("id", "platform", "zernio_account_id", "username", "is_active", "synced_at", "created_at")

	for _, a := range accounts {
		query = query.Values(
			a.GetID().String(),
			string(a.GetPlatform()),
			a.GetZernioAccountID(),
			a.GetUsername(),
			a.IsActive(),
			a.GetSyncedAt(),
			a.GetCreatedAt(),
		)
	}

	sqlQuery, args, err := query.
		Suffix(`ON CONFLICT (zernio_account_id) DO UPDATE SET
			platform = EXCLUDED.platform,
			username = EXCLUDED.username,
			is_active = EXCLUDED.is_active,
			synced_at = EXCLUDED.synced_at`).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlQuery, args...)
	return err
}

func (s *storage) DeactivateMissing(ctx context.Context, presentZernioIDs []string) error {
	query := squirrel.Update(accountsTableName).
		Set("is_active", false).
		Set("synced_at", time.Now().UTC()).
		Where(squirrel.Eq{"is_active": true})

	if len(presentZernioIDs) > 0 {
		query = query.Where(squirrel.NotEq{"zernio_account_id": presentZernioIDs})
	}

	sqlQuery, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlQuery, args...)
	return err
}
