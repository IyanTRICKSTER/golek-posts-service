package database

import (
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMongoDB(t *testing.T) {

	//Load .Env
	err := godotenv.Load("../../.env")
	if err != nil {
		panic(err)
	}

	db := Database{
		DbName:       os.Getenv("DB_NAME"),
		DbCollection: os.Getenv("DB_COLLECTION"),
		DbHost:       os.Getenv("DB_HOST"),
		DbPort:       os.Getenv("DB_PORT"),
		DbUsername:   os.Getenv("DB_USERNAME"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
	}
	db.Prepare()

	assert.NotEqual(t, db.GetConnection(), nil)

}
