package services

import (
	"context"
	"golek_posts_service/cmd/msg_broker"
	"golek_posts_service/pkg/contracts/status"
	"golek_posts_service/pkg/database"
	"golek_posts_service/pkg/http/requests"
	"golek_posts_service/pkg/models"
	"golek_posts_service/pkg/repositories"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestPostService(t *testing.T) {

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

	postRepository := repositories.NewPostRepository(db.GetConnection(), db.GetCollection())
	qrcodeRepository := repositories.NewQRCodeRepository()
	awsS3Repository := repositories.NewS3Repository(
		os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
		os.Getenv("AWS_BUCKET_NAME"),
		os.Getenv("AWS_BUCKET_REGION"),
		[]string{"image/jpeg", "image/png"},
		int64(5*1024*1024), //MAX Filesize 5mb
		3,
	)

	//Establish Message Broker Connection
	amqpConn := msg_broker.New(
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)
	mqPublisherService := msg_broker.NewMQPublisher(amqpConn)
	mqPublisherService.Setup()

	postService := NewPostService(&postRepository, &qrcodeRepository, &awsS3Repository, &mqPublisherService)

	var createdPostID string

	t.Run("Create", func(t *testing.T) {
		createdPost, opStatus, err := postService.Create(context.TODO(), requests.CreatePostRequest{

			Title: "Samsung A35",
			//ImageURL:           "https://randomwordgenerator.com/img/picture-generator/57e8dc4a4c57a914f1dc8460962e33791c3ad6e04e50744172287edc964dc6_640.jpg",
			Place:           "Lantai 2 depan ruang dosen",
			Description:     "",
			Characteristics: []requests.PostCharacteristicRequest{{Title: "Casing hitam"}, {"Ada gantungan boneka"}},
		})
		if err != nil || opStatus == status.PostCreatedStatusFailed {
			t.Error(err)
		}

		assert.Equal(t, createdPost.Title, "Samsung A35")
		assert.Equal(t, len(createdPost.Characteristics), 2)

		createdPostID = createdPost.ID.Hex()
	})

	t.Run("RequestValidateOwner", func(t *testing.T) {
		qrcodeUrl, opStatus, err := postService.RequestValidateOwner(context.TODO(), createdPostID)
		if err != nil || opStatus == status.PostRequestValidationSuccess {
			t.Error(opStatus, err)
		}

		t.Log(qrcodeUrl)
	})

	t.Run("ValidateOwner", func(t *testing.T) {
		opStatus, err := postService.ValidateOwner(context.TODO(), requests.ValidateItemOwnerRequest{
			PostID:  "636f342ce0ebfaa96cfef711",
			OwnerID: 112,
			Hash:    "$2a$10$VdTa0yoL2SPlUzyxpOkgcu5FFIx0IOhWt4au2WbmQ.NXhABZVExee",
		})
		if err != nil || opStatus != status.PostValidateOwnerSuccess {
			t.Error(err)
		}

		assert.Truef(t, opStatus == status.PostValidateOwnerSuccess, "Validation Owner success, item has returned")
	})

	t.Run("Search", func(t *testing.T) {
		paginate := models.Pagination{
			Page:    1,
			PerPage: 25,
		}
		posts, err := postService.Search(context.TODO(), "i", paginate)
		if err != nil {
			t.Error(err.Error())
		}

		assert.IsType(t, []models.Post{}, posts)
	})

}
