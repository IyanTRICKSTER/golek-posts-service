package requests

import "mime/multipart"

type CreatePostRequest struct {
	//UserID          int64                       `binding:"required" form:"user_id" json:"user_id"`
	Title           string                      `binding:"required" form:"title"`
	Image           *multipart.FileHeader       `binding:"required" form:"image" `
	Place           string                      `binding:"required" form:"place"`
	Description     string                      `binding:"" form:"description"`
	Characteristics []PostCharacteristicRequest `binding:"required" form:"characteristics"`
}

type UpdatePostRequest struct {
	//UserID          int64                       `form:"user_id" json:"user_id"`
	//ReturnedTo      int64                       ` form:"returned_to" json:"returned_to,omitempty"`
	Title           string                      `binding:"required" form:"title" json:"title"`
	Image           *multipart.FileHeader       `binding:"" form:"image" json:"image"`
	Place           string                      `binding:"required" form:"place" json:"place"`
	Description     string                      `binding:"" form:"description" json:"description,omitempty"`
	Characteristics []PostCharacteristicRequest `binding:"required" form:"characteristics" json:"characteristics"`
}

type ValidateItemOwnerRequest struct {
	PostID string `json:"post_id" binding:"required"`
	Hash   string `json:"hash" binding:"required"`
}

type PostCharacteristicRequest struct {
	Title string `binding:"required" form:"title" json:"title"`
}
