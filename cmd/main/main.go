package main

import (
	"golek_posts_service/cmd/grpc_server"
	"golek_posts_service/cmd/msg_broker"
	"golek_posts_service/pkg/contracts"
	"golek_posts_service/pkg/database"
	"golek_posts_service/pkg/database/migration"
	"golek_posts_service/pkg/http/controllers"
	"golek_posts_service/pkg/repositories"
	"golek_posts_service/pkg/services"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
		DbPort:       os.Getenv("DB_PORT_IN"),
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

	//Establish Message Broker Connection
	amqpConn := msg_broker.New(
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)
	mqPublisherService := msg_broker.NewMQPublisher(amqpConn)
	mqPublisherService.Setup()

	//Initialize Services
	postService := services.NewPostService(&postRepository, &qrcodeRepository, &awsS3Repository, &mqPublisherService)

	//Initialize Routes
	controllers.SetupHandler(engine, &postService)

	//Run grpc service in different thread
	go func(postService *contracts.PostServiceContract) {
		grpcServer := grpc_server.New(postService)
		grpcServer.Run()
	}(&postService)

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
