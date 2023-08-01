package contracts

import (
	"context"
	"github.com/aws/aws-sdk-go/service/s3"
	"golek_posts_service/pkg/http/responses"
	"mime/multipart"
)

type StorageRepository interface {
	UploadFiles(files []*multipart.FileHeader, prefix string) ([]responses.S3Response, error)
	UploadFile(file *multipart.FileHeader, prefix string) (responses.S3Response, error)
	DeleteObject(objectKey *string) error
	GetClient() (*s3.S3, error)
	ReadFileBytes(file *multipart.FileHeader) ([]byte, error)
}

type ICloudStorageRepo interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, name string) (string, error)
	DeleteFile(ctx context.Context, fileUrl string) error
	Connect() error
}
