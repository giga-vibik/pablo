package api

import (
	"errors"
	"io"
	"log"
	"net/http"

	mediaService "github.com/pablo/backend/internal/service/media"
)

// maxVideoSize — потолок на загружаемое видео. Reels/Shorts всё равно короткие,
// а без лимита один запрос может выесть всю память.
const maxVideoSize = 512 << 20 // 512 MB

func (s *Server) UploadVideo(w http.ResponseWriter, r *http.Request, postId string) {
	ctx := r.Context()

	postID, err := parsePostID(postId)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post_id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxVideoSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read file")
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "video/mp4"
	}

	m, err := s.mediaService.UploadVideo(ctx, postID, header.Filename, mimeType, content)
	if err != nil {
		if errors.Is(err, mediaService.ErrEmptyFile) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Println("error: while uploading video", err.Error())
		writeError(w, http.StatusInternalServerError, "failed to upload video")
		return
	}

	writeJSON(w, http.StatusOK, toSchemaMedia(m))
}
