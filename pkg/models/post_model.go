package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Post struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	UserID          int64              `bson:"user_id" json:"user_id"`
	ReturnedTo      int64              `bson:"returned_to" json:"returned_to,omitempty"`
	IsReturned      bool               `bson:"is_returned" json:"is_returned"`
	Title           string             `bson:"title" json:"title"`
	ImageURL        string             `bson:"image_url" json:"image_url"`
	ImageKey        string             `bson:"image_key"`
	ConfirmationKey string             `bson:"confirmation_key" json:"-"`
	Place           string             `bson:"place" json:"place"`
	Description     string             `bson:"description" json:"description,omitempty"`
	Characteristics []Characteristic   `bson:"characteristics" json:"characteristics"`
	UpdatedAt       *time.Time         `json:"updated_at,omitempty" bson:"updated_at"`
	CreatedAt       *time.Time         `json:"created_at,omitempty" bson:"created_at"`
	DeletedAt       *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at"`
}

type Characteristic struct {
	Title string `bson:"title" json:"title"`
}

func (p *Post) SetPostID(postID primitive.ObjectID) {
	p.ID = postID
}

type Pagination struct {
	Page    int64
	PerPage int64
}

func (p Pagination) GetPagination() (limit int64, skip int64) {
	return p.PerPage, (p.Page - 1) * p.PerPage
}
