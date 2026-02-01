package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email         string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Username      string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	PasswordHash  string         `gorm:"type:varchar(255);not null" json:"-"`
	Role          string         `gorm:"type:varchar(50);not null;default:'user'" json:"role"`
	Status        string         `gorm:"type:varchar(50);not null;default:'active'" json:"status"`
	EmailVerified bool           `gorm:"not null;default:false" json:"email_verified"`
	LastLoginAt   *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt     time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	OAuthAccounts  []OAuthAccount  `gorm:"foreignKey:UserID" json:"oauth_accounts,omitempty"`
	APIKeys        []UserAPIKey    `gorm:"foreignKey:UserID" json:"api_keys,omitempty"`
	PaymentRecords []PaymentRecord `gorm:"foreignKey:UserID" json:"payment_records,omitempty"`
	BillingRecords []BillingRecord `gorm:"foreignKey:UserID" json:"billing_records,omitempty"`
}

type OAuthProvider struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name         string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	ClientID     string    `gorm:"type:varchar(255);not null" json:"client_id"`
	ClientSecret string    `gorm:"type:varchar(512);not null" json:"-"`
	Config       JSONB     `gorm:"type:jsonb;not null;default:'{}'" json:"config"`
	Enabled      bool      `gorm:"not null;default:true" json:"enabled"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`

	OAuthAccounts []OAuthAccount `gorm:"foreignKey:ProviderID" json:"oauth_accounts,omitempty"`
}

type OAuthAccount struct {
	ID             string     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID         string     `gorm:"type:uuid;not null;index" json:"user_id"`
	ProviderID     string     `gorm:"type:uuid;not null;index" json:"provider_id"`
	ProviderUserID string     `gorm:"type:varchar(255);not null" json:"provider_user_id"`
	AccessToken    string     `gorm:"type:varchar(512)" json:"-"`
	RefreshToken   string     `gorm:"type:varchar(512)" json:"-"`
	TokenExpiresAt *time.Time `json:"token_expires_at,omitempty"`
	CreatedAt      time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null" json:"updated_at"`

	User     User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Provider OAuthProvider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
}

type JSONB map[string]interface{}

func (j JSONB) GormDataType() string {
	return "jsonb"
}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return json.Unmarshal([]byte(fmt.Sprintf("%v", v)), j)
	}
	return json.Unmarshal(data, j)
}

func (User) TableName() string {
	return "users"
}

func (OAuthProvider) TableName() string {
	return "oauth_providers"
}

func (OAuthAccount) TableName() string {
	return "oauth_accounts"
}
