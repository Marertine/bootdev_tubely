package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

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

	myMediaType, _, err := mime.ParseMediaType(mediaType)
	if err != nil || (myMediaType != "image/jpeg" && myMediaType != "image/png") {
		respondWithError(w, http.StatusBadRequest, "Thumbnail must be an image of type jpeg or png", err)
		return
	}

	fileData := make([]byte, header.Size)
	fileData, err = io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't read thumbnail file", err)
		return
	}

	/*// Create a 32-byte slice
	sliceByte := make([]byte, 32)

	// Fill it with cryptographically secure random data
	if _, err := rand.Read(sliceByte); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create cryptographically secure random data", err)
		return
	}

	// Encode using base64 URL encoding without padding
	randomString := base64.RawURLEncoding.EncodeToString(sliceByte)*/

	randomFilename, err := helperReturn32RandomChars()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create random filename", err)
		return
	}

	//myFilepath := filepath.Join(cfg.assetsRoot, fmt.Sprintf("%s%s", videoID.String(), filepath.Ext(header.Filename)))
	//myFilepath := filepath.Join(cfg.assetsRoot, fmt.Sprintf("%s%s", randomString, filepath.Ext(header.Filename)))
	myFilepath := filepath.Join(cfg.assetsRoot, fmt.Sprintf("%s%s", randomFilename, filepath.Ext(header.Filename)))
	myfile, err := os.Create(myFilepath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create thumbnail file", err)
		return
	}
	defer myfile.Close()

	_, err = myfile.Write(fileData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't write thumbnail file", err)
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

	// Set the URL for the thumbnail
	//dataURL := fmt.Sprintf("http://%s/assets/%s%s", r.Host, videoID.String(), filepath.Ext(header.Filename))
	//dataURL := fmt.Sprintf("http://%s/assets/%s%s", r.Host, randomString, filepath.Ext(header.Filename))
	dataURL := fmt.Sprintf("http://%s/assets/%s%s", r.Host, randomFilename, filepath.Ext(header.Filename))
	db_video.ThumbnailURL = &dataURL

	err = cfg.db.UpdateVideo(db_video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't UpdateVideo", err)
		return
	}

	respondWithJSON(w, http.StatusOK, db_video)
}
