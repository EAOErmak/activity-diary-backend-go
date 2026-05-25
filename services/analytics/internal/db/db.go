package db

import (
	"database/sql"
	"time"

	"activitydiary/analytics/internal/config"

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
