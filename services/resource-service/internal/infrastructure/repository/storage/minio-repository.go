package storage

import (
	"context"
	"time"

	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioInstance struct {
	Client     *minio.Client
	bucketName string
}

func NewMinioInstance(endpoint string, accessKeyID string, secretAccessKey string, useSSL bool, bucketName string) (*minioInstance, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &minioInstance{
		Client:     minioClient,
		bucketName: bucketName,
	}, nil
}

func (r *minioInstance) EnsureBucketExists(ctx context.Context) error {
	err := r.Client.MakeBucket(ctx, r.bucketName, minio.MakeBucketOptions{Region: "us-east-1"})
	if err != nil {
		exists, errBucketExists := r.Client.BucketExists(ctx, r.bucketName)
		if errBucketExists != nil && !exists {
			return err
		}
	}
	return nil
}

func (r *minioInstance) GeneratePSUrl(ctx context.Context, resourceID string, objectKey string, expireIn int64) (*types.UploadUrlResponse, error) {
	// Generate the Presigned PUT URL

	uploadURL, err := r.Client.PresignedPutObject(ctx, r.bucketName, objectKey, time.Duration(expireIn)*time.Second)
	if err != nil {
		return nil, err
	}
	return &types.UploadUrlResponse{
		UploadUrl:  uploadURL.String(),
		ObjectKey:  objectKey,
		ExpiresIn:  expireIn,
		ResourceID: resourceID,
	}, nil
}
