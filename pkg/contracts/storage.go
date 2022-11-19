package contracts

import (
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

type StorageService interface {
	UploadFiles(files []*multipart.FileHeader, prefix string) ([]responses.S3Response, error)
	UploadFile(file *multipart.FileHeader, prefix string) (responses.S3Response, error)
	Delete(objectKey string) error
}
