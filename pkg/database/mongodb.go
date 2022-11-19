package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golek_posts_service/pkg/contracts"
	"log"
)

type Database struct {
	DbUsername   string
	DBPassword   string
	DbName       string
	DbHost       string
	DbPort       string
	DbCollection string
	collection   *mongo.Collection
	connection   *mongo.Database
}

func (db *Database) Prepare() contracts.MongoDBContract {

	//defer catchError()

	if db.connection == nil {

		clientOptions := options.Client().ApplyURI(db.Dsn())

		//clientOptions := options.Client().
		//	ApplyURI(fmt.Sprintf("mongodb://%s:%s", db.DbHost, db.DbPort)).
		//	SetAuth(options.Credential{
		//		AuthSource:    "admin",
		//		AuthMechanism: "SCRAM-SHA-256",
		//		Username:      db.DbUsername,
		//		Password:      db.DBPassword,
		//	})

		client, err := mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			log.Println("Create Mongo NewClient Error")
			panic(err.Error())
		}

		log.Println("PINGING: MongoDB")
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			panic(err.Error())
		}

		log.Println("Connected to the database: MongoDB")

		db.connection = client.Database(db.DbName)

	} else {
		log.Println("Already Connected to the database: MongoDB")
	}

	return db
}

func (db *Database) GetConnection() *mongo.Database {
	return db.connection
}

func (db *Database) GetCollection() *mongo.Collection {
	return db.connection.Collection(db.DbCollection)
}

func (db *Database) Dsn() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?", db.DbUsername, db.DBPassword, db.DbHost, db.DbPort, db.DbName)
	//return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin&ssl=false", db.DbUsername, db.DBPassword, db.DbHost, db.DbPort, db.DbName)
}

func catchError() {
	if r := recover(); r != nil {
		log.Println("Some Error Occurred During Database Establishment", r)
	}
}
