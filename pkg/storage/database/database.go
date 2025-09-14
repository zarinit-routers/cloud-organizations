package database

import (
	"fmt"

	"github.com/charmbracelet/log"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Migrate() error {
	source := migrate.FileMigrationSource{
		Dir: "migrations",
	}
	conn, err := connect()
	if err != nil {
		return fmt.Errorf("failed connect to database: %s", err)
	}

	db, _ := conn.DB()
	count, err := migrate.Exec(db, "postgres", source, migrate.Up)
	if err != nil {
		log.Error("Failed migrate database", "error", err)
		return err
	}
	log.Info("Migrated database", "migrations", count)
	return nil
}

func Connect() (*gorm.DB, error) {
	conn, err := connect()
	if err != nil {
		log.Error("Failed connect database", "error", err)
		return nil, fmt.Errorf("failed connect to database: %s", err)
	}
	return conn, nil
}

func MustConnect() *gorm.DB {
	conn, err := connect()
	if err != nil {
		log.Fatal("Failed connect database", "error", err)
	}
	return conn
}
func connect() (*gorm.DB, error) {
	connectionString := viper.GetString("db-connection-string")
	if connectionString == "" {
		return nil, fmt.Errorf("database connection string is not set")
	}
	db, err := gorm.Open(postgres.Open(connectionString))
	if err != nil {
		return nil, err
	}
	return db, nil
}
