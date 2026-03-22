package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var s3Client *s3.S3
var bucketName string

const (
	MaxImageSize = 5 << 20 // 5 MB
)

var (
	ValidImageExtensions = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	ValidImageMimeTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
)

func InitializeStorage() error {
	// Validate S3/Storage credentials
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		log.Fatal("AWS_ACCESS_KEY_ID environment variable must be set")
	}
	if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Fatal("AWS_SECRET_ACCESS_KEY environment variable must be set")
	}
	if os.Getenv("S3_ENDPOINT") == "" {
		log.Fatal("S3_ENDPOINT environment variable must be set")
	}
	if os.Getenv("S3_BUCKET") == "" {
		log.Fatal("S3_BUCKET environment variable must be set")
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")
	endpoint := os.Getenv("S3_ENDPOINT")
	bucketName = os.Getenv("S3_BUCKET")

	if accessKey == "" || secretKey == "" || endpoint == "" || bucketName == "" {
		return fmt.Errorf("AWS credentials and S3_ENDPOINT must be set")
	}

	if region == "" {
		region = "ap-southeast-1"
	}

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	s3Client = s3.New(sess)
	return nil
}

func ValidateImageFile(fileHeader *multipart.FileHeader) ([]byte, error) {
	ext := strings.ToLower(path.Ext(fileHeader.Filename))
	if !ValidImageExtensions[ext] {
		return nil, fmt.Errorf("invalid file extension: only .jpg, .jpeg, .png, .gif, .webp allowed")
	}

	if fileHeader.Size > MaxImageSize {
		return nil, fmt.Errorf("file is too large: max %d MB per file", MaxImageSize/(1<<20))
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	detectedType := http.DetectContentType(fileBytes)

	if !ValidImageMimeTypes[detectedType] {
		return nil, fmt.Errorf("invalid file type: detected %s, only image files (jpg, png, gif, webp) are allowed", detectedType)
	}

	return fileBytes, nil
}

func UploadImage(fileHeader *multipart.FileHeader) (string, error) {
	fileBytes, err := ValidateImageFile(fileHeader)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(fileBytes)

	ext := strings.ToLower(path.Ext(fileHeader.Filename))
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	filename := hex.EncodeToString(randomBytes) + ext

	putObjectInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
		ACL:         aws.String("public-read"),
	}

	_, err = s3Client.PutObject(putObjectInput)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	endpoint := os.Getenv("S3_ENDPOINT")
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", endpoint, bucketName, filename)

	return publicURL, nil
}

func DeleteImage(photoURL string) error {
	if photoURL == "" {
		return nil
	}

	// Extract filename from URL
	// URL format: https://xxxxx.supabase.co/storage/v1/object/public/cafe-photos/abc123.jpg
	parts := strings.Split(photoURL, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid photo URL format")
	}
	filename := parts[len(parts)-1] // Get last part (the filename)

	deleteObjectInput := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	}

	_, err := s3Client.DeleteObject(deleteObjectInput)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}
