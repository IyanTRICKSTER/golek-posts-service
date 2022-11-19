package migration

import "golek_posts_service/pkg/database"

type Migration struct {
	DB *database.Database
}

func NewMigration(db *database.Database) *Migration {
	return &Migration{DB: db}
}
