package contracts

import (
	"context"
	"golek_posts_service/pkg/contracts/status"
	"golek_posts_service/pkg/http/requests"
	"golek_posts_service/pkg/models"
)

type PostServiceContract interface {
	Fetch(ctx context.Context, pagination models.Pagination) ([]models.Post, error)
	FindById(ctx context.Context, postID string) (models.Post, error)
	Search(ctx context.Context, keyword string, pagination models.Pagination) ([]models.Post, error)
	Create(ctx context.Context, request requests.CreatePostRequest) (models.Post, status.PostOperationStatus, error)
	Update(ctx context.Context, postID string, request requests.UpdatePostRequest) (models.Post, status.PostOperationStatus, error)
	Delete(ctx context.Context, postID string) (status.PostOperationStatus, error)
	RequestValidateOwner(ctx context.Context, postID string) (qrCode string, status status.PostOperationStatus, err error)
	ValidateOwner(ctx context.Context, request requests.ValidateItemOwnerRequest) (status.PostOperationStatus, error)
}

type PostRepositoryContract interface {
	Fetch(ctx context.Context, latest bool, limit int64, skip int64) ([]models.Post, error)
	FindById(ctx context.Context, postID string) (models.Post, error)
	Search(ctx context.Context, keyword string, limit int64, skip int64) ([]models.Post, error)
	Create(ctx context.Context, post models.Post) (models.Post, status.PostOperationStatus, error)
	Update(ctx context.Context, postID string, post models.Post) (models.Post, status.PostOperationStatus, error)
	Delete(ctx context.Context, postID string) (status.PostOperationStatus, error)
}
