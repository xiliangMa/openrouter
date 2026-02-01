package model

import (
	"time"

	"gorm.io/gorm"
)

type ModelStatistic struct {
	ID              string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ModelID         string    `gorm:"type:uuid;not null;index" json:"model_id"`
	Date            time.Time `gorm:"type:date;not null;index" json:"date"`
	TotalRequests   int       `gorm:"not null;default:0" json:"total_requests"`
	TotalTokens     int       `gorm:"not null;default:0" json:"total_tokens"`
	AvgResponseTime float64   `gorm:"type:decimal(10,3);not null;default:0" json:"avg_response_time"`
	SuccessRate     float64   `gorm:"type:decimal(5,2);not null;default:0" json:"success_rate"`
	CreatedAt       time.Time `gorm:"not null" json:"created_at"`

	Model Model `gorm:"foreignKey:ModelID" json:"model,omitempty"`
}

func (ModelStatistic) TableName() string {
	return "model_statistics"
}

type MigrationRecord struct {
	gorm.Model
	Name  string `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	Batch int    `gorm:"not null" json:"batch"`
}

func (MigrationRecord) TableName() string {
	return "migrations"
}
