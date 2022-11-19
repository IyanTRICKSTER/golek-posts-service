package repositories

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"golek_posts_service/pkg/contracts"
	"golek_posts_service/pkg/database"
	"golek_posts_service/pkg/database/migration"
	"golek_posts_service/pkg/http/controllers"
	"golek_posts_service/pkg/http/requests"
	"golek_posts_service/pkg/services"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	engine *gin.Engine
	s3Repo contracts.StorageRepository
)

func init() {

	//Load .Env
	err := godotenv.Load("../../.env")
	if err != nil {
		panic(err)
	}

	//Create Gin Instance
	engine = gin.Default()

	db := database.Database{
		DbName:       os.Getenv("DB_NAME"),
		DbCollection: os.Getenv("DB_COLLECTION"),
		DbHost:       os.Getenv("DB_HOST"),
		DbPort:       os.Getenv("DB_PORT"),
		DbUsername:   os.Getenv("DB_USERNAME"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
	}
	db.Prepare()

	//Migrate DB
	m := migration.NewMigration(&db)
	m.MigrateSettings()

	//Initialize Repositories
	postRepository := NewPostRepository(db.GetConnection(), db.GetCollection())
	qrcodeRepository := NewQRCodeRepository()
	s3Repo = NewS3Repository(
		os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
		os.Getenv("AWS_BUCKET_NAME"),
		os.Getenv("AWS_BUCKET_REGION"),
		[]string{"image/jpeg", "image/png", "image/jpg"},
		int64(5*1024*1024), //MAX Filesize 5mb
		3,
	)

	//Initialize Services
	postService := services.NewPostService(&postRepository, &qrcodeRepository, &s3Repo)

	//Initialize Routes
	controllers.SetupHandler(engine, &postService)
}

func TestAwsS3StorageRepository(t *testing.T) {

	reqBody := new(bytes.Buffer)

	writer := multipart.NewWriter(reqBody)

	formField := map[string]string{
		"user_id":         "100",
		"title":           "Post Test",
		"place":           "dekat kantor polisi",
		"characteristics": `{"title": "warna biru"}`,
	}

	formFiles := map[string]string{
		"image": "../statics/testpost1.jpeg",
	}

	for k, v := range formField {
		err := writer.WriteField(k, v)
		if err != nil {
			t.Error(err)
		}
	}

	for k, v := range formFiles {
		file, err := os.Open(v)
		if err != nil {
			t.Error(err)
		}

		w, err := writer.CreateFormFile(k, v)
		if err != nil {
			t.Error(err)
		}

		if _, err := io.Copy(w, file); err != nil {
			t.Error(err)
		}
	}

	// Close the writer
	err := writer.Close()
	if err != nil {
		t.Error(err)
	}

	t.Run("ReadFileBytes-MultipartHeader", func(t *testing.T) {

		engine.POST("/readimage/", func(context *gin.Context) {
			var createReq requests.CreatePostRequest
			err := context.ShouldBind(&createReq)
			if err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			fileBytes, err := s3Repo.ReadFileBytes(createReq.Image)
			if err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			context.JSON(http.StatusOK, gin.H{
				"files_bytes": fileBytes,
			})

		})

		httpReq, _ := http.NewRequest("POST", "/readimage/", reqBody)
		httpReq.Header.Add("Content-Type", writer.FormDataContentType())
		httpReq.Header.Add("X-User-Id", "100")
		httpReq.Header.Add("X-User-Role", "user")
		httpReq.Header.Add("X-User-Permission", "c")

		//Listen Result
		httpRes := httptest.NewRecorder()

		//Proceed Request
		engine.ServeHTTP(httpRes, httpReq)

		//Read Result
		response, _ := ioutil.ReadAll(httpRes.Body)

		assert.NotEqual(t, http.StatusInternalServerError, httpRes.Code)
		assert.NotEmpty(t, response)

	})

	t.Run("CreatePost", func(t *testing.T) {

		//Prepare Request
		httpReq, _ := http.NewRequest("POST", "/api/posts/", reqBody)
		httpReq.Header.Add("Accept", "*/*")
		httpReq.Header.Add("Content-Type", writer.FormDataContentType())
		httpReq.Header.Add("X-User-Id", "100")
		httpReq.Header.Add("X-User-Role", "user")
		httpReq.Header.Add("X-User-Permission", "c")

		//Listen Result
		httpRes := httptest.NewRecorder()

		//Proceed Request
		engine.ServeHTTP(httpRes, httpReq)

		//Read Result
		response, _ := ioutil.ReadAll(httpRes.Body)

		t.Log(string(response))
		assert.Equal(t, http.StatusCreated, httpRes.Code)

	})

	t.Run("UpdatePost", func(t *testing.T) {
		//Prepare Request
		httpReq, _ := http.NewRequest("PUT", "/api/posts/637675e15d056f482217bb9e", reqBody)
		httpReq.Header.Add("Accept", "*/*")
		httpReq.Header.Add("Content-Type", writer.FormDataContentType())
		httpReq.Header.Add("X-User-Id", "100")
		httpReq.Header.Add("X-User-Role", "user")
		httpReq.Header.Add("X-User-Permission", "u")

		//Listen Result
		httpRes := httptest.NewRecorder()

		//Proceed Request
		engine.ServeHTTP(httpRes, httpReq)

		//Read Result
		response, _ := ioutil.ReadAll(httpRes.Body)

		t.Log(string(response))
		assert.Equal(t, http.StatusOK, httpRes.Code)
	})

}
