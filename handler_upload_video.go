package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
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

	db_video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't GetVideo", err)
		return
	}

	if db_video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "You do not own this video", nil)
		return
	}

	fmt.Println("uploading video", videoID, "by user", userID)

	const uploadLimit = 1 << 30 // 1 GB
	r.Body = http.MaxBytesReader(w, r.Body, uploadLimit)

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get video from form", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}

	parsedMediaType, _, err := mime.ParseMediaType(mediaType)
	if err != nil || (parsedMediaType != "video/mp4") {
		respondWithError(w, http.StatusBadRequest, "Video must be of type mp4", err)
		return
	}

	myTempFile, err := os.CreateTemp("", "tubely-temp-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create temp file", err)
		return
	}

	defer os.Remove(myTempFile.Name())
	defer myTempFile.Close()

	_, err = io.Copy(myTempFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't write video file", err)
		return
	}

	// Must reset the file pointer after copying to allow reading the file again from the beginning
	_, err = myTempFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't reset file pointer", err)
		return
	}

	randomFilename, err := helperReturn32RandomChars()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create random filename", err)
		return
	}

	myS3Bucket := cfg.s3Bucket
	myS3Key := fmt.Sprintf("%s%s", randomFilename, filepath.Ext(header.Filename))

	myPutObjectParams := s3.PutObjectInput{
		Bucket:      &myS3Bucket,
		Key:         &myS3Key,
		Body:        myTempFile,
		ContentType: &parsedMediaType,
	}

	// Save the data to a file on Amazon S3
	_, err = cfg.s3Client.PutObject(r.Context(), &myPutObjectParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't upload video to S3", err)
		return
	}

	// Set the URL for the video
	dataURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, myS3Key)
	db_video.VideoURL = &dataURL

	err = cfg.db.UpdateVideo(db_video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't UpdateVideo", err)
		return
	}

	respondWithJSON(w, http.StatusOK, db_video)
}
