package storage

import (
	"context"
	"time"

	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioInstance struct {
	Client        *minio.Client
	presignClient *minio.Client
	bucketName    string
}

func NewMinioInstance(endpoint, publicEndpoint, accessKeyID, secretAccessKey string, useSSL bool, bucketName string) (*minioInstance, error) {
	internalClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
		Region: "us-east-1",
	})
	if err != nil {
		return nil, err
	}

	presignClient, err := minio.New(publicEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
		Region: "us-east-1", // Disables internal network probe for bucket region
	})
	if err != nil {
		return nil, err
	}

	return &minioInstance{
		Client:        internalClient,
		presignClient: presignClient,
		bucketName:    bucketName,
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
	// Generate the Presigned PUT URL using the public endpoint for browser client uploads
	uploadURL, err := r.presignClient.PresignedPutObject(ctx, r.bucketName, objectKey, time.Duration(expireIn)*time.Second)
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
