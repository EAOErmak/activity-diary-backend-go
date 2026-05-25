package db

import (
	"database/sql"
	"fmt"
	"time"

	"activitydiary/api/internal/config"
	"activitydiary/api/internal/models"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(cfg config.Config) (*gorm.DB, *sql.DB, error) {
	gormDB, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, nil, err
	}

	return gormDB, sqlDB, nil
}

func RunMigrations(cfg config.Config) error {
	m, err := migrate.New(fmt.Sprintf("file://%s", cfg.MigrationsPath), cfg.DatabaseURL())
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func SeedInitialData(database *gorm.DB) error {
	if err := seedAdmin(database); err != nil {
		return err
	}
	return seedReferenceData(database)
}

func seedAdmin(database *gorm.DB) error {
	var count int64
	if err := database.Model(&models.User{}).Where("email = ?", "admin@example.com").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return database.Create(&models.User{
		Email:        "admin@example.com",
		Username:     "admin",
		FullName:     "Admin User",
		PasswordHash: string(hash),
		Role:         "ADMIN",
		Enabled:      true,
	}).Error
}

func seedReferenceData(database *gorm.DB) error {
	names := []models.DictionaryItem{
		{Type: "METRIC_NAME", Label: "pull ups", Active: true},
		{Type: "METRIC_NAME", Label: "push ups", Active: true},
		{Type: "METRIC_UNIT", Label: "reps", Active: true},
		{Type: "METRIC_UNIT", Label: "kg", Active: true},
		{Type: "CATEGORY", Label: "food", Active: true},
		{Type: "SUB_CATEGORY", Label: "protein", Active: true},
	}
	for _, item := range names {
		if err := database.Where("type = ? AND label = ?", item.Type, item.Label).FirstOrCreate(&models.DictionaryItem{}, item).Error; err != nil {
			return err
		}
	}

	var metricName, metricUnit models.DictionaryItem
	if err := database.Where("type = ? AND label = ?", "METRIC_NAME", "pull ups").First(&metricName).Error; err != nil {
		return err
	}
	if err := database.Where("type = ? AND label = ?", "METRIC_UNIT", "reps").First(&metricUnit).Error; err != nil {
		return err
	}
	if err := database.Where("name = ?", "training").FirstOrCreate(&models.Tag{}, models.Tag{Name: "training", Status: "APPROVED"}).Error; err != nil {
		return err
	}
	var tag models.Tag
	if err := database.Where("name = ?", "training").First(&tag).Error; err != nil {
		return err
	}
	if err := database.Where("metric_name_id = ? AND metric_unit_id = ?", metricName.ID, metricUnit.ID).
		FirstOrCreate(&models.MetricNameUnitLink{}, models.MetricNameUnitLink{MetricNameID: metricName.ID, MetricUnitID: metricUnit.ID}).Error; err != nil {
		return err
	}
	if err := database.Where("tag_id = ? AND metric_name_id = ?", tag.ID, metricName.ID).
		FirstOrCreate(&models.TagMetricLink{}, models.TagMetricLink{TagID: tag.ID, MetricNameID: metricName.ID}).Error; err != nil {
		return err
	}
	return database.Where("tag_id = ? AND chart_type = ?", tag.ID, "TRAINING_RAW").
		FirstOrCreate(&models.TagChartTypeLink{}, models.TagChartTypeLink{TagID: tag.ID, ChartType: "TRAINING_RAW"}).Error
}
