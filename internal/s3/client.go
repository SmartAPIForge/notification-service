package s3

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"net/http"
	"notification-service/internal/config"
	"os"
)

type S3Client struct {
	client *s3.S3
	bucket string
}

func NewS3Client(cfg *config.Config) *S3Client {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("eu-central-1"),
		Credentials: credentials.NewStaticCredentials(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Endpoint:    aws.String(cfg.S3Path),
	})
	if err != nil {
		return nil
	}

	return &S3Client{
		client: s3.New(sess),
		bucket: cfg.S3Bucket,
	}
}

func (r *S3Client) LoadFile(url string) (*os.File, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file from URL %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch file from URL %s: status %s", url, resp.Status)
	}

	tempFile, err := os.CreateTemp("", "s3file-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to write data to temporary file: %v", err)
	}

	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to seek temporary file: %v", err)
	}

	return tempFile, nil
}
