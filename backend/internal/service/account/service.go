package account

import (
	"context"
	"log"

	accountDomain "github.com/pablo/backend/internal/domain/account"
	postDomain "github.com/pablo/backend/internal/domain/post"
	"github.com/pablo/backend/internal/integration/zernio"
)

type AccountStorage interface {
	ListAccounts(ctx context.Context) ([]accountDomain.Account, error)
	UpsertAccounts(ctx context.Context, accounts []accountDomain.Account) error
	DeactivateMissing(ctx context.Context, presentZernioIDs []string) error
}

type ZernioClient interface {
	IsConfigured() bool
	ListAccounts(ctx context.Context) ([]zernio.Account, error)
	GetConnectURL(ctx context.Context, platform, redirectURL string) (string, error)
}

type AccountService interface {
	// ListAccounts отдаёт локальный кэш аккаунтов.
	ListAccounts(ctx context.Context) ([]accountDomain.Account, error)
	// SyncAccounts подтягивает актуальный список из zernio.
	SyncAccounts(ctx context.Context) ([]accountDomain.Account, error)
	// GetConnectURL возвращает ссылку hosted-OAuth для подключения площадки.
	GetConnectURL(ctx context.Context, platform postDomain.Platform, redirectURL string) (string, error)
}

type service struct {
	accountStorage AccountStorage
	zernioClient   ZernioClient
}

func NewAccountService(accountStorage AccountStorage, zernioClient ZernioClient) AccountService {
	return &service{
		accountStorage: accountStorage,
		zernioClient:   zernioClient,
	}
}

func (s *service) ListAccounts(ctx context.Context) ([]accountDomain.Account, error) {
	return s.accountStorage.ListAccounts(ctx)
}

func (s *service) SyncAccounts(ctx context.Context) ([]accountDomain.Account, error) {
	if !s.zernioClient.IsConfigured() {
		return nil, zernio.ErrNotConfigured
	}

	remote, err := s.zernioClient.ListAccounts(ctx)
	if err != nil {
		log.Println("error: while listing zernio accounts", err.Error())
		return nil, err
	}

	accounts := make([]accountDomain.Account, 0, len(remote))
	presentIDs := make([]string, 0, len(remote))

	for _, r := range remote {
		isActive := r.Status == "" || r.Status == "connected"
		accounts = append(accounts, accountDomain.NewAccount(
			postDomain.Platform(r.Platform),
			r.ID,
			r.Username,
			isActive,
		))
		presentIDs = append(presentIDs, r.ID)
	}

	if err = s.accountStorage.UpsertAccounts(ctx, accounts); err != nil {
		return nil, err
	}

	// Аккаунт, отключённый в zernio, пропадает из выдачи — гасим его локально,
	// иначе фронт продолжит предлагать площадку, в которую публикация упадёт.
	if err = s.accountStorage.DeactivateMissing(ctx, presentIDs); err != nil {
		return nil, err
	}

	return s.accountStorage.ListAccounts(ctx)
}

func (s *service) GetConnectURL(
	ctx context.Context,
	platform postDomain.Platform,
	redirectURL string,
) (string, error) {
	if !s.zernioClient.IsConfigured() {
		return "", zernio.ErrNotConfigured
	}

	return s.zernioClient.GetConnectURL(ctx, string(platform), redirectURL)
}
