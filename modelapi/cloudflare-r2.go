package modelapi

import (
	"bytes"
	"context"
	"errors"
	"fairytale-creator/logger"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Uploader 封装了Cloudflare R2上传和预签名URL生成功能
type R2Uploader struct {
	client     *s3.Client
	bucketName string
}

// NewR2Uploader 创建一个新的R2上传器实例
func NewR2Uploader(accountID, accessKeyID, accessKeySecret, bucketName string) (*R2Uploader, error) {
	// 配置AWS SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	// 创建S3客户端并设置自定义端点
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
	})

	return &R2Uploader{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// UploadFromLocalFile 从本地文件上传到R2
func (u *R2Uploader) UploadFromLocalFile(localFilePath, objectKey string) error {
	// 打开本地文件
	file, err := os.Open(localFilePath)
	if err != nil {
		logger.Error("failed to open local file: " + err.Error())
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer file.Close()

	// 获取文件信息以检查文件大小等（可选）
	fileInfo, err := file.Stat()
	if err != nil {
		logger.Error("failed to get file info: " + err.Error())
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// 根据文件扩展名获取内容类型
	contentType := u.getContentTypeFromExtension(objectKey)

	// 上传到R2
	_, err = u.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(u.bucketName),
		Key:           aws.String(objectKey),
		Body:          file,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(fileInfo.Size()), // 可选，但推荐提供
	})
	if err != nil {
		logger.Error("failed to upload to R2: " + err.Error())
		return fmt.Errorf("failed to upload to R2: %v", err)
	}

	logger.Log("Successfully uploaded %s to R2 bucket %s", objectKey, u.bucketName)
	return nil
}

// UploadFromURL 从图片URL下载并上传到R2
func (u *R2Uploader) UploadFromURL(imageURL, objectKey string) error {
	// 下载图片
	resp, err := http.Get(imageURL)
	if err != nil {
		logger.Error("failed to download image: " + err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("failed to download image, status code: " + strconv.Itoa(resp.StatusCode))
		return errors.New("failed to download image, status code: " + strconv.Itoa(resp.StatusCode))
	}

	// 获取内容类型
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// 根据文件扩展名猜测内容类型
		contentType = u.getContentTypeFromExtension(objectKey)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read response body: " + err.Error())
		return err
	}

	// 上传到R2
	output, err := u.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(u.bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		logger.Error("failed to upload to R2: " + err.Error())
		return err
	}

	logger.Log("Successfully uploaded %s to R2 bucket %s", objectKey, u.bucketName)
	logger.Log(*output.ChecksumCRC32)
	return nil
}

// GeneratePresignedURL 生成预签名URL
func (u *R2Uploader) GeneratePresignedURL(objectKey string, expiration time.Duration) (string, error) {
	// 创建预签名客户端
	presignClient := s3.NewPresignClient(u.client)

	// 创建获取对象输入
	input := &s3.GetObjectInput{
		Bucket: aws.String(u.bucketName),
		Key:    aws.String(objectKey),
	}

	// 生成预签名URL
	presignedRequest, err := presignClient.PresignGetObject(context.TODO(), input, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return presignedRequest.URL, nil
}

// getContentTypeFromExtension 根据文件扩展名返回内容类型
func (u *R2Uploader) getContentTypeFromExtension(filename string) string {
	ext := path.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
