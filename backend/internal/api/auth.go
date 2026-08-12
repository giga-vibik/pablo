package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/pablo/backend/internal/auth"
	"github.com/pablo/backend/schema"
)

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req schema.LoginJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	token, err := s.authManager.Login(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid login or password")
			return
		}

		log.Println("error: while generating token", err.Error())
		writeError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	writeJSON(w, http.StatusOK, struct {
		AccessToken string `json:"access_token"`
	}{AccessToken: token})
}
