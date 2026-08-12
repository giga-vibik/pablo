package api

import (
	"errors"
	"log"
	"net/http"

	postDomain "github.com/pablo/backend/internal/domain/post"
	"github.com/pablo/backend/internal/integration/zernio"
	"github.com/pablo/backend/schema"
)

type accountsResponse struct {
	Accounts []schema.Account `json:"accounts"`
}

func (s *Server) ListAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accounts, err := s.accountService.ListAccounts(ctx)
	if err != nil {
		log.Println("error: while listing accounts", err.Error())
		writeError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}

	writeJSON(w, http.StatusOK, accountsResponse{Accounts: toSchemaAccounts(accounts)})
}

func (s *Server) SyncAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accounts, err := s.accountService.SyncAccounts(ctx)
	if err != nil {
		log.Println("error: while syncing accounts", err.Error())
		writeError(w, http.StatusInternalServerError, "failed to sync accounts")
		return
	}

	writeJSON(w, http.StatusOK, accountsResponse{Accounts: toSchemaAccounts(accounts)})
}

func (s *Server) GetConnectURL(w http.ResponseWriter, r *http.Request, params schema.GetConnectURLParams) {
	ctx := r.Context()

	platform := postDomain.Platform(params.Platform)
	if !platform.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid platform")
		return
	}

	redirectURL := ""
	if params.RedirectUrl != nil {
		redirectURL = *params.RedirectUrl
	}

	authURL, err := s.accountService.GetConnectURL(ctx, platform, redirectURL)
	if err != nil {
		if errors.Is(err, zernio.ErrNotConfigured) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Println("error: while getting connect url", err.Error())
		writeError(w, http.StatusInternalServerError, "failed to get connect url")
		return
	}

	writeJSON(w, http.StatusOK, struct {
		AuthURL string `json:"auth_url"`
	}{AuthURL: authURL})
}
