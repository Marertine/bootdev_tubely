package main

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {

	presignedClient := s3.NewPresignClient(s3Client)

	presignedURL, err := presignedClient.PresignGetObject(
		context.Background(),
		&s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &key,
		},
		s3.WithPresignExpires(expireTime),
	)
	if err != nil {
		return "", err
	}

	return presignedURL.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {

	// Make sure we actually have a VideoURL
	if video.VideoURL == nil {
		// Nothing to sign, return as-is
		return video, nil
	}

	strVideo := strings.Split(*video.VideoURL, ",")
	// Make sure we have a Bucket and a Key, separated by a comma
	if len(strVideo) < 2 {
		// Malformed value, return as-is
		return video, nil
	}

	signedURL, err := generatePresignedURL(cfg.s3Client, strVideo[0], strVideo[1], (300 * time.Second))
	if err != nil {
		return video, err
	}

	video.VideoURL = &signedURL
	return video, nil
}
