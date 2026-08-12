package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pablo/backend/internal/auth"
	"github.com/pablo/backend/internal/service"
	"github.com/pablo/backend/internal/service/account"
	"github.com/pablo/backend/internal/service/media"
	"github.com/pablo/backend/internal/service/post"
	"github.com/pablo/backend/schema"
)

type Server struct {
	postService    post.PostService
	mediaService   media.MediaService
	accountService account.AccountService
	authManager    auth.AuthManager
}

func NewHttpServer(serviceRegistry *service.Services, authManager auth.AuthManager) schema.ServerInterface {
	return &Server{
		postService:    serviceRegistry.PostService,
		mediaService:   serviceRegistry.MediaService,
		accountService: serviceRegistry.AccountService,
		authManager:    authManager,
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if body == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Println("error: while writing response", err.Error())
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, schema.Error{Message: message})
}
