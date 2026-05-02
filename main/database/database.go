package database

import (
	"fmt"
	"os"
	"strings"

	"go-learn/main/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	dsn := databaseDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.DictionaryItem{},
		&models.DiaryEntry{},
		&models.EntryMetric{},
		&models.EntryMetricValue{},
	); err != nil {
		return err
	}

	if err := dropLegacyIndexes(db); err != nil {
		return err
	}

	DB = db
	return nil
}

func databaseDSN() string {
	host := envOrDefault("DB_HOST", "localhost")
	port := envOrDefault("DB_PORT", "5432")
	user := envOrDefault("DB_USER", "postgres")
	password := envOrDefault("DB_PASSWORD", "postgres")
	name := envOrDefault("DB_NAME", "postgres")
	sslMode := envOrDefault("DB_SSLMODE", "disable")

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host,
		user,
		password,
		name,
		port,
		sslMode,
	)
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func dropLegacyIndexes(db *gorm.DB) error {
	indexes := []struct {
		model any
		name  string
	}{
		{model: &models.DictionaryItem{}, name: "udx_dictionary_type_label"},
		{model: &models.DiaryEntry{}, name: "idx_diary_started"},
		{model: &models.EntryMetric{}, name: "idx_metric_type"},
		{model: &models.EntryMetric{}, name: "idx_metric_entry"},
		{model: &models.EntryMetricValue{}, name: "udx_metric_unit"},
	}

	for _, index := range indexes {
		if !db.Migrator().HasIndex(index.model, index.name) {
			continue
		}

		if err := db.Migrator().DropIndex(index.model, index.name); err != nil {
			return err
		}
	}

	return nil
}
