package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	maxMemory := int64(10 << 20) // 10 MB
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse multipart form", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get thumbnail from form", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}

	fileData := make([]byte, header.Size)
	fileData, err = io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't read thumbnail file", err)
		return
	}

	db_video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't GetVideo", err)
		return
	}

	if db_video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "You do not own this video", nil)
		return
	}

	videoThumbnails[videoID] = thumbnail{
		data:      fileData,
		mediaType: mediaType,
	}

	/*thumbURL := fmt.Sprintf("%s/thumbnails/%s%s", cfg.s3CfDistribution, videoID.String(), filepath.Ext(header.Filename))
	err = cfg.db.UpdateVideoThumbnailURL(videoID, thumbURL)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't UpdateVideoThumbnailURL", err)
		return
	}*/
	thumbURL := fmt.Sprintf("http://localhost:%v/api/thumbnails/%s", cfg.port, videoID.String())
	db_video.ThumbnailURL = &thumbURL

	err = cfg.db.UpdateVideo(db_video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't UpdateVideo", err)
		return
	}

	respondWithJSON(w, http.StatusOK, db_video)
}
