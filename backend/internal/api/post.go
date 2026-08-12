package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"

	postDomain "github.com/pablo/backend/internal/domain/post"
	postService "github.com/pablo/backend/internal/service/post"
	"github.com/pablo/backend/schema"
)

const (
	defaultPostsLimit = 50
	maxPostsLimit     = 200
)

func (s *Server) ListPosts(w http.ResponseWriter, r *http.Request, params schema.ListPostsParams) {
	ctx := r.Context()

	limit := defaultPostsLimit
	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	if limit > maxPostsLimit {
		limit = maxPostsLimit
	}

	offset := 0
	if params.Offset != nil && *params.Offset > 0 {
		offset = *params.Offset
	}

	posts, err := s.postService.ListPosts(ctx, limit, offset)
	if err != nil {
		log.Println("error: while listing posts", err.Error())
		writeError(w, http.StatusInternalServerError, "failed to list posts")
		return
	}

	res := make([]schema.Post, 0, len(posts))
	for i := range posts {
		res = append(res, toSchemaPost(posts[i].Post, posts[i].Media))
	}

	writeJSON(w, http.StatusOK, struct {
		Posts []schema.Post `json:"posts"`
	}{Posts: res})
}

func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req schema.CreatePostJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	kind := postDomain.Kind(req.Kind)
	if kind != postDomain.VideoKind && kind != postDomain.TextKind {
		writeError(w, http.StatusBadRequest, "invalid kind")
		return
	}

	if len(req.Targets) == 0 {
		writeError(w, http.StatusBadRequest, "targets are required")
		return
	}

	content := ""
	if req.Content != nil {
		content = *req.Content
	}

	post := postDomain.NewPost(kind, content, req.ScheduledAt)

	targets := make([]postDomain.Target, 0, len(req.Targets))
	for _, t := range req.Targets {
		platform := postDomain.Platform(t.Platform)
		if !platform.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid platform: "+string(t.Platform))
			return
		}

		caption := ""
		if t.Caption != nil {
			caption = *t.Caption
		}

		targets = append(targets, postDomain.NewTarget(post.GetID(), platform, caption))
	}

	if err := s.postService.CreatePost(ctx, post, targets); err != nil {
		if errors.Is(err, postService.ErrTextPlatform) || errors.Is(err, postService.ErrNoTargets) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Println("error: while creating post", err.Error())
		writeError(w, http.StatusInternalServerError, "failed to create post")
		return
	}

	post.SetTargets(targets)

	writeJSON(w, http.StatusOK, toSchemaPost(post, nil))
}

func (s *Server) GetPost(w http.ResponseWriter, r *http.Request, postId string) {
	ctx := r.Context()

	postID, err := parsePostID(postId)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post_id")
		return
	}

	post, media, err := s.postService.GetPost(ctx, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "post not found")
			return
		}

		log.Println("error: while getting post", err.Error())
		writeError(w, http.StatusInternalServerError, "failed to get post")
		return
	}

	writeJSON(w, http.StatusOK, toSchemaPost(post, media))
}

func (s *Server) DeletePost(w http.ResponseWriter, r *http.Request, postId string) {
	ctx := r.Context()

	postID, err := parsePostID(postId)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post_id")
		return
	}

	if err = s.postService.DeletePost(ctx, postID); err != nil {
		log.Println("error: while deleting post", err.Error())
		writeError(w, http.StatusInternalServerError, "failed to delete post")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) PublishPost(w http.ResponseWriter, r *http.Request, postId string) {
	ctx := r.Context()

	postID, err := parsePostID(postId)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post_id")
		return
	}

	post, err := s.postService.PublishPost(ctx, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "post not found")
			return
		}
		if errors.Is(err, postService.ErrNoVideo) || errors.Is(err, postService.ErrNoTargets) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Println("error: while publishing post", err.Error())
		writeError(w, http.StatusInternalServerError, "failed to publish post")
		return
	}

	writeJSON(w, http.StatusOK, toSchemaPost(post, nil))
}

func parsePostID(raw string) (postDomain.PostID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return postDomain.PostID{}, err
	}

	return postDomain.PostID(id), nil
}
