package model

import (
	"time"
)

type PaymentRecord struct {
	ID            string     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID        string     `gorm:"type:uuid;not null;index" json:"user_id"`
	Amount        float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	Currency      string     `gorm:"type:varchar(10);not null;default:'CNY'" json:"currency"`
	PaymentMethod string     `gorm:"type:varchar(50);not null" json:"payment_method"`
	TransactionID string     `gorm:"type:varchar(255);uniqueIndex" json:"transaction_id,omitempty"`
	Status        string     `gorm:"type:varchar(50);not null;default:'pending';index" json:"status"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	Metadata      JSONB      `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt     time.Time  `gorm:"not null;index" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null" json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type BillingRecord struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID         string    `gorm:"type:uuid;not null;index" json:"user_id"`
	APIKeyID       *string   `gorm:"type:uuid;index" json:"api_key_id,omitempty"`
	ModelID        string    `gorm:"type:uuid;not null;index" json:"model_id"`
	RequestTokens  int       `gorm:"not null;default:0" json:"request_tokens"`
	ResponseTokens int       `gorm:"not null;default:0" json:"response_tokens"`
	TotalTokens    int       `gorm:"not null;default:0" json:"total_tokens"`
	Cost           float64   `gorm:"type:decimal(10,8);not null;default:0" json:"cost"`
	Metadata       JSONB     `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt      time.Time `gorm:"not null;index" json:"created_at"`

	User   User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	APIKey UserAPIKey `gorm:"foreignKey:APIKeyID" json:"api_key,omitempty"`
	Model  Model      `gorm:"foreignKey:ModelID" json:"model,omitempty"`
}

type SystemConfig struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Key         string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"key"`
	Value       string    `gorm:"type:text" json:"value,omitempty"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	IsPublic    bool      `gorm:"not null;default:false" json:"is_public"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`
}

func (PaymentRecord) TableName() string {
	return "payment_records"
}

func (BillingRecord) TableName() string {
	return "billing_records"
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
