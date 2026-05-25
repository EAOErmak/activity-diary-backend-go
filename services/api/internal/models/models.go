package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Email        string    `gorm:"uniqueIndex;size:255;not null"`
	Username     string    `gorm:"uniqueIndex;size:255;not null"`
	FullName     string    `gorm:"size:255;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	Role         string    `gorm:"size:32;not null"`
	Enabled      bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	TokenHash string    `gorm:"size:255;not null"`
	Revoked   bool      `gorm:"not null;default:false"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

type DiaryEntry struct {
	ID          uint `gorm:"primaryKey"`
	UserID      uint `gorm:"index;not null"`
	WhenStarted *time.Time
	WhenEnded   *time.Time
	Duration    *int
	Mood        *int
	Description *string
	Status      string `gorm:"size:32;index;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Metrics     []DiaryEntryMetric `gorm:"foreignKey:DiaryEntryID"`
	Tags        []DiaryEntryTag    `gorm:"foreignKey:DiaryEntryID"`
}

type DiaryEntryMetric struct {
	ID           uint `gorm:"primaryKey"`
	DiaryEntryID uint `gorm:"index;not null"`
	MetricTypeID uint `gorm:"index;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Values       []DiaryEntryMetricValue `gorm:"foreignKey:DiaryEntryMetricID"`
}

type DiaryEntryMetricValue struct {
	ID                 uint    `gorm:"primaryKey"`
	DiaryEntryMetricID uint    `gorm:"index;not null"`
	UnitID             uint    `gorm:"index;not null"`
	Value              float64 `gorm:"type:numeric;not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Tag struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex;size:255;not null"`
	Status    string `gorm:"size:32;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DiaryEntryTag struct {
	DiaryEntryID uint `gorm:"primaryKey"`
	TagID        uint `gorm:"primaryKey"`
}

type DictionaryItem struct {
	ID          uint    `gorm:"primaryKey"`
	Type        string  `gorm:"size:32;index;not null"`
	Label       string  `gorm:"size:255;index;not null"`
	Active      bool    `gorm:"not null;default:true"`
	AllowedRole *string `gorm:"size:32"`
	ParentID    *uint
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MetricNameUnitLink struct {
	MetricNameID uint `gorm:"primaryKey"`
	MetricUnitID uint `gorm:"primaryKey"`
}

type TagMetricLink struct {
	TagID        uint `gorm:"primaryKey"`
	MetricNameID uint `gorm:"primaryKey"`
}

type TagChartTypeLink struct {
	TagID     uint   `gorm:"primaryKey"`
	ChartType string `gorm:"primaryKey;size:64"`
}

type GeneralFood struct {
	ID               uint    `gorm:"primaryKey"`
	DictionaryItemID uint    `gorm:"uniqueIndex;not null"`
	Protein          float64 `gorm:"type:numeric;not null"`
	Fat              float64 `gorm:"type:numeric;not null"`
	Carbs            float64 `gorm:"type:numeric;not null"`
	Callories        float64 `gorm:"type:numeric;not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
