package video

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioInstance struct {
	Client     *minio.Client
	bucketName string
}

func NewMinioInstance(endpoint string, accessKeyID string, secretAccessKey string, useSSL bool, bucketName string) (*MinioInstance, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &MinioInstance{
		Client:     minioClient,
		bucketName: bucketName,
	}, nil
}

func (r *MinioInstance) Download(ctx context.Context, objectName string, localFilePath string) error {

	// Download the file directly to your local system
	return r.Client.FGetObject(ctx, r.bucketName, objectName, localFilePath, minio.GetObjectOptions{})

}

func (r *MinioInstance) Upload(ctx context.Context, objectName string, localFilePath string, contentType string) error {
	_, err := r.Client.FPutObject(ctx, r.bucketName, objectName, localFilePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return err
	}
	return nil
}
