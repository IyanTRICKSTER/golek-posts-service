package repositories

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golek_posts_service/pkg/contracts/status"
	"golek_posts_service/pkg/database"
	"golek_posts_service/pkg/models"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestDatabaseRepository(t *testing.T) {

	//Load .Env
	err := godotenv.Load("../../.env")
	if err != nil {
		panic(err)
	}

	db := database.Database{
		DbName:       os.Getenv("DB_NAME"),
		DbCollection: os.Getenv("DB_COLLECTION"),
		DbHost:       os.Getenv("DB_HOST"),
		DbPort:       os.Getenv("DB_PORT"),
		DbUsername:   os.Getenv("DB_USERNAME"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
	}
	db.Prepare()
	dbRepo := NewPostRepository(db.GetConnection(), db.GetCollection())

	t.Run("Fetch", func(t *testing.T) {
		posts, err := dbRepo.Fetch(context.TODO(), false, 10, 0, map[string]any{})
		if err != nil {
			t.Error(err)
		}
		t.Log(posts)
	})

	t.Run("FetchById", func(t *testing.T) {
		t.Run("Post Not Exist", func(t *testing.T) {
			_, err := dbRepo.FindById(context.TODO(), primitive.NewObjectID().Hex())
			assert.Equal(t, err, mongo.ErrNoDocuments)
		})
	})

	t.Run("Create", func(t *testing.T) {

		timeNow := time.Now()

		post := models.Post{
			ID:              primitive.NewObjectID(),
			UserID:          rand.Int63n(99999),
			ReturnedTo:      0,
			IsReturned:      false,
			Title:           "Post 3",
			ImageURL:        "https://randomwordgenerator.com/img/picture-generator/57e8dc4a4c57a914f1dc8460962e33791c3ad6e04e50744172287edc964dc6_640.jpg",
			Place:           "Near enter gates",
			Description:     "",
			Characteristics: []models.Characteristic{{"1. Warna merah"}, {"2. Gores sudut kanan atas"}},
			UpdatedAt:       &timeNow,
			CreatedAt:       &timeNow,
			DeletedAt:       nil,
		}

		createdPost, opStatus, err := dbRepo.Create(context.TODO(), post)
		if err != nil {
			t.Error(err)
		}

		if opStatus == status.PostCreatedStatusFailed {
			t.Error("Failed create new post")
		}

		assert.Equal(t, post, createdPost)
	})

	t.Run("Update", func(t *testing.T) {

		timeNow := time.Now()
		postID := primitive.NewObjectID()

		//1. Create Post
		post := models.Post{
			ID:              postID,
			UserID:          rand.Int63n(99999),
			ReturnedTo:      rand.Int63n(99999),
			IsReturned:      true,
			Title:           "Post 4",
			ImageURL:        "https://randomwordgenerator.com/img/picture-generator/57e8dc4a4c57a914f1dc8460962e33791c3ad6e04e50744172287edc964dc6_640.jpg",
			Place:           "Near enter gates",
			Description:     "",
			Characteristics: []models.Characteristic{{"1. Warna merah"}, {"2. Gores sudut kanan atas"}},
			UpdatedAt:       &timeNow,
			CreatedAt:       &timeNow,
			DeletedAt:       nil,
		}

		createdPost, opStatus, err := dbRepo.Create(context.TODO(), post)
		if err != nil {
			t.Error(err)
		}

		if opStatus == status.PostCreatedStatusFailed {
			t.Error("Failed create new post")
		}

		//2. Update Post Title
		createdPost.Title = "New title of post 4"

		updatedPost, updatedStatus, err := dbRepo.Update(context.TODO(), postID.Hex(), createdPost)
		if err != nil {
			t.Error(err)
		}

		if updatedStatus == status.PostUpdatedStatusFailed {
			t.Error("Failed to update post 4")
		}

		assert.Equal(t, createdPost.Title, updatedPost.Title)
	})

	t.Run("Delete", func(t *testing.T) {
		deletedStatus, err := dbRepo.Delete(context.TODO(), "6373836eacb626cfb10ece78")
		if err != nil {
			t.Error(err)
		}

		if deletedStatus == status.PostDeletedStatusFailed {
			t.Error("Delete Post Failed")
		}

		assert.Equal(t, deletedStatus, status.PostDeletedStatusSuccess)
	})

	t.Run("Search", func(t *testing.T) {
		posts, err := dbRepo.Search(context.TODO(), "dokumen", 3, 1)
		if err != nil {
			t.Log(err)
		}

		for _, d := range posts {
			log.Println(d)
		}
		assert.IsType(t, []models.Post{}, posts)
	})
}
