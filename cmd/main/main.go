package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golek_posts_service/pkg/database"
	"golek_posts_service/pkg/database/migration"
	"golek_posts_service/pkg/http/controllers"
	"golek_posts_service/pkg/repositories"
	"golek_posts_service/pkg/services"
	"os"
)

func main() {

	//Load .Env
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	//Create Gin Instance
	engine := gin.Default()

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

	//Initialize Services
	postService := services.NewPostService(&postRepository, &qrcodeRepository, &awsS3Repository)

	//Initialize Routes
	controllers.SetupHandler(engine, &postService)

	//Running App With Desired Port
	if port := os.Getenv("APP_PORT"); port == "" {

		err := engine.Run(":8080")
		if err != nil {
			panic(err)
		}

	} else {

		err := engine.Run(":" + port)
		if err != nil {
			panic(err)
		}
	}

}
