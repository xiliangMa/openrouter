package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
	"massrouter.ai/backend/pkg/utils"
)

func main() {
	dsn := "host=localhost port=5432 user=massrouter password=changeme dbname=massrouter sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	ctx := context.Background()

	fmt.Println("Seeding model providers...")
	providers := []*model.ModelProvider{
		{Name: "OpenAI", APIBaseURL: "https://api.openai.com/v1", APIKey: os.Getenv("OPENAI_API_KEY"), Config: model.JSONB{"description": "OpenAI API provider"}, Status: "active"},
		{Name: "Anthropic", APIBaseURL: "https://api.anthropic.com", APIKey: os.Getenv("ANTHROPIC_API_KEY"), Config: model.JSONB{"description": "Anthropic Claude API"}, Status: "active"},
		{Name: "Google", APIBaseURL: "https://generativelanguage.googleapis.com", APIKey: os.Getenv("GOOGLE_API_KEY"), Config: model.JSONB{"description": "Google Gemini API"}, Status: "active"},
		{Name: "Meta", APIBaseURL: "https://api.meta.ai", APIKey: os.Getenv("META_API_KEY"), Config: model.JSONB{"description": "Meta Llama API"}, Status: "active"},
		{Name: "Cohere", APIBaseURL: "https://api.cohere.ai", APIKey: os.Getenv("COHERE_API_KEY"), Config: model.JSONB{"description": "Cohere API"}, Status: "active"},
	}

	for _, p := range providers {
		var existing model.ModelProvider
		if err := db.WithContext(ctx).Where("name = ?", p.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.WithContext(ctx).Create(p).Error; err != nil {
					log.Printf("Failed to create provider %s: %v", p.Name, err)
				} else {
					fmt.Printf("✅ Created provider: %s\n", p.Name)
				}
			}
		} else {
			// Update existing provider with API key if not set
			updateNeeded := false
			if p.APIKey != "" && existing.APIKey != p.APIKey {
				existing.APIKey = p.APIKey
				updateNeeded = true
			}
			if updateNeeded {
				if err := db.WithContext(ctx).Save(&existing).Error; err != nil {
					log.Printf("Failed to update provider %s: %v", p.Name, err)
				} else {
					fmt.Printf("✅ Updated provider API key: %s\n", p.Name)
				}
			} else {
				fmt.Printf("⚠️ Provider exists: %s\n", p.Name)
			}
		}
	}

	// Get provider IDs
	var providerMap = make(map[string]string)
	for _, p := range providers {
		var prov model.ModelProvider
		if err := db.WithContext(ctx).Where("name = ?", p.Name).First(&prov).Error; err == nil {
			providerMap[p.Name] = prov.ID
		}
	}

	intPtr := func(i int) *int { return &i }

	fmt.Println("\nSeeding models...")
	models := []*model.Model{
		// OpenAI
		{Name: "gpt-4o", Description: "OpenAI GPT-4 Omni", ProviderID: providerMap["OpenAI"], ContextLength: intPtr(128000), MaxTokens: intPtr(4096), InputPrice: 0.005, OutputPrice: 0.015, IsActive: true, IsFree: false, Category: "chat", Capabilities: model.JSONB{"vision": true}, PricingTier: "premium"},
		{Name: "gpt-3.5-turbo", Description: "OpenAI GPT-3.5 Turbo", ProviderID: providerMap["OpenAI"], ContextLength: intPtr(16385), MaxTokens: intPtr(4096), InputPrice: 0.0005, OutputPrice: 0.0015, IsActive: true, IsFree: false, Category: "chat", Capabilities: model.JSONB{}, PricingTier: "standard"},
		// Anthropic
		{Name: "claude-3-sonnet", Description: "Anthropic Claude 3 Sonnet", ProviderID: providerMap["Anthropic"], ContextLength: intPtr(200000), MaxTokens: intPtr(4096), InputPrice: 0.003, OutputPrice: 0.015, IsActive: true, IsFree: false, Category: "chat", Capabilities: model.JSONB{"vision": true}, PricingTier: "standard"},
		{Name: "claude-3-haiku", Description: "Anthropic Claude 3 Haiku", ProviderID: providerMap["Anthropic"], ContextLength: intPtr(200000), MaxTokens: intPtr(4096), InputPrice: 0.00025, OutputPrice: 0.00125, IsActive: true, IsFree: false, Category: "chat", Capabilities: model.JSONB{"vision": true}, PricingTier: "economy"},
		// Google
		{Name: "gemini-1.5-pro", Description: "Google Gemini 1.5 Pro", ProviderID: providerMap["Google"], ContextLength: intPtr(1000000), MaxTokens: intPtr(8192), InputPrice: 0.00125, OutputPrice: 0.00375, IsActive: true, IsFree: false, Category: "chat", Capabilities: model.JSONB{"vision": true}, PricingTier: "premium"},
		// Meta
		{Name: "llama-3-70b", Description: "Meta Llama 3 70B", ProviderID: providerMap["Meta"], ContextLength: intPtr(8192), MaxTokens: intPtr(4096), InputPrice: 0.0009, OutputPrice: 0.0009, IsActive: true, IsFree: false, Category: "chat", Capabilities: model.JSONB{}, PricingTier: "standard"},
		// Cohere
		{Name: "command-r", Description: "Cohere Command R", ProviderID: providerMap["Cohere"], ContextLength: intPtr(128000), MaxTokens: intPtr(4096), InputPrice: 0.0005, OutputPrice: 0.0015, IsActive: true, IsFree: false, Category: "chat", Capabilities: model.JSONB{}, PricingTier: "standard"},
	}

	for _, m := range models {
		var existing model.Model
		if err := db.WithContext(ctx).Where("name = ? AND provider_id = ?", m.Name, m.ProviderID).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.WithContext(ctx).Create(m).Error; err != nil {
					log.Printf("Failed to create model %s: %v", m.Name, err)
				} else {
					fmt.Printf("✅ Created model: %s\n", m.Name)
				}
			}
		} else {
			fmt.Printf("⚠️ Model exists: %s\n", m.Name)
		}
	}

	fmt.Println("\nSeeding system configs...")
	configs := []*model.SystemConfig{
		{Key: "site_name", Value: "MassRouter SaaS", Description: "Website name", IsPublic: true},
		{Key: "default_user_balance", Value: "10.0", Description: "Default balance for new users", IsPublic: false},
		{Key: "rate_limit_per_minute", Value: "60", Description: "API rate limit per minute per user", IsPublic: false},
		{Key: "maintenance_mode", Value: "false", Description: "Maintenance mode toggle", IsPublic: true},
		{Key: "registration_enabled", Value: "true", Description: "User registration enabled", IsPublic: true},
	}

	for _, c := range configs {
		var existing model.SystemConfig
		if err := db.WithContext(ctx).Where("key = ?", c.Key).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.WithContext(ctx).Create(c).Error; err != nil {
					log.Printf("Failed to create config %s: %v", c.Key, err)
				} else {
					fmt.Printf("✅ Created config: %s\n", c.Key)
				}
			}
		} else {
			fmt.Printf("⚠️ Config exists: %s\n", c.Key)
		}
	}

	fmt.Println("\nSeeding admin user...")
	// Check for existing admin users with either email
	adminEmails := []string{"admin@massrouter.ai", "admin@openrouter.ai"}

	for _, adminEmail := range adminEmails {
		var existingAdmin model.User
		if err := db.WithContext(ctx).Where("email = ?", adminEmail).First(&existingAdmin).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue // Try next email
			}
		} else {
			// Found existing admin user
			fmt.Printf("⚠️ Admin user already exists: %s (role: %s)\n", existingAdmin.Email, existingAdmin.Role)
			// Update role to admin if not already
			if existingAdmin.Role != "admin" {
				existingAdmin.Role = "admin"
				if err := db.WithContext(ctx).Save(&existingAdmin).Error; err != nil {
					log.Printf("Failed to update admin user role: %v", err)
				} else {
					fmt.Printf("✅ Updated user role to admin: %s\n", existingAdmin.Email)
				}
			}
			// Also update username to 'admin' if different
			if existingAdmin.Username != "admin" {
				existingAdmin.Username = "admin"
				if err := db.WithContext(ctx).Save(&existingAdmin).Error; err != nil {
					log.Printf("Failed to update admin username: %v", err)
				}
			}
			// Update password to admin123 if needed
			if !utils.VerifyPassword("admin123", existingAdmin.PasswordHash) {
				newHash, err := utils.HashPassword("admin123")
				if err != nil {
					log.Printf("Failed to hash admin password: %v", err)
				} else {
					existingAdmin.PasswordHash = newHash
					if err := db.WithContext(ctx).Save(&existingAdmin).Error; err != nil {
						log.Printf("Failed to update admin password: %v", err)
					} else {
						fmt.Printf("✅ Updated admin password to admin123 for: %s\n", existingAdmin.Email)
					}
				}
			}
			break
		}
	}

	// If no admin user found, create one
	var adminUserCount int64
	db.WithContext(ctx).Model(&model.User{}).Where("username = ? OR email IN (?)", "admin", adminEmails).Count(&adminUserCount)
	if adminUserCount == 0 {
		// Hash password for admin user
		passwordHash, err := utils.HashPassword("admin123")
		if err != nil {
			log.Printf("Failed to hash admin password: %v", err)
		} else {
			adminUser := &model.User{
				Email:         "admin@massrouter.ai",
				Username:      "admin",
				PasswordHash:  passwordHash,
				Role:          "admin",
				Status:        "active",
				EmailVerified: true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			if err := db.WithContext(ctx).Create(adminUser).Error; err != nil {
				log.Printf("Failed to create admin user: %v", err)
			} else {
				fmt.Printf("✅ Created admin user: %s (password: admin123)\n", adminUser.Email)
			}
		}
	}

	fmt.Println("\nSeeding OAuth providers...")
	oauthProviders := []*model.OAuthProvider{
		{
			Name:         "github",
			ClientID:     "",
			ClientSecret: "",
			Config:       model.JSONB{"display_name": "GitHub", "auth_url": "https://github.com/login/oauth/authorize", "token_url": "https://github.com/login/oauth/access_token", "user_info_url": "https://api.github.com/user", "scopes": "user:email"},
			Enabled:      true,
		},
		{
			Name:         "google",
			ClientID:     "",
			ClientSecret: "",
			Config:       model.JSONB{"display_name": "Google", "auth_url": "https://accounts.google.com/o/oauth2/auth", "token_url": "https://oauth2.googleapis.com/token", "user_info_url": "https://www.googleapis.com/oauth2/v3/userinfo", "scopes": "profile email"},
			Enabled:      true,
		},
	}

	for _, p := range oauthProviders {
		var existing model.OAuthProvider
		if err := db.WithContext(ctx).Where("name = ?", p.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.WithContext(ctx).Create(p).Error; err != nil {
					log.Printf("Failed to create OAuth provider %s: %v", p.Name, err)
				} else {
					fmt.Printf("✅ Created OAuth provider: %s\n", p.Name)
				}
			}
		} else {
			fmt.Printf("⚠️ OAuth provider exists: %s\n", p.Name)
		}
	}

	// Get existing user IDs
	var users []model.User
	db.WithContext(ctx).Find(&users)
	if len(users) == 0 {
		fmt.Println("⚠️ No users found, skipping billing data")
	} else {
		fmt.Println("\nSeeding payment records...")
		// Create payment records for each user
		for i, user := range users {
			if i >= 3 { // Limit to first 3 users
				break
			}
			paymentMethods := []string{"wechat", "alipay", "credit_card", "bank_transfer"}
			statuses := []string{"completed", "pending", "failed"}

			for j := 0; j < 3; j++ { // 3 payments per user
				paidAt := time.Now().Add(-time.Duration(j*24) * time.Hour)
				status := statuses[j%len(statuses)]
				var paidAtPtr *time.Time
				if status == "completed" {
					paidAtPtr = &paidAt
				}

				payment := &model.PaymentRecord{
					UserID:        user.ID,
					Amount:        float64((j+1)*50 + i*100), // Varying amounts
					Currency:      "CNY",
					PaymentMethod: paymentMethods[(i+j)%len(paymentMethods)],
					TransactionID: fmt.Sprintf("txn_%s_%d_%d", user.ID[:8], i, j),
					Status:        status,
					PaidAt:        paidAtPtr,
					Metadata:      model.JSONB{"note": "Test payment"},
				}

				var existing model.PaymentRecord
				if err := db.WithContext(ctx).Where("transaction_id = ?", payment.TransactionID).First(&existing).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						if err := db.WithContext(ctx).Create(payment).Error; err != nil {
							log.Printf("Failed to create payment record: %v", err)
						} else {
							fmt.Printf("✅ Created payment: %s for user %s\n", payment.TransactionID, user.Email)
						}
					}
				} else {
					fmt.Printf("⚠️ Payment exists: %s\n", payment.TransactionID)
				}
			}
		}

		fmt.Println("\nSeeding billing records...")
		// Get models and API keys
		var models []model.Model
		db.WithContext(ctx).Find(&models)

		var apiKeys []model.UserAPIKey
		db.WithContext(ctx).Find(&apiKeys)

		// Create billing records for each user
		for i, user := range users {
			if i >= 2 { // Limit to first 2 users
				break
			}

			// Create 5-10 billing records per user
			for j := 0; j < 5+int(user.ID[0])%5; j++ {
				modelIdx := j % len(models)
				apiKeyID := ""
				if len(apiKeys) > 0 {
					apiKeyID = apiKeys[0].ID
				}

				requestTokens := 100 + j*50
				responseTokens := 200 + j*30
				totalTokens := requestTokens + responseTokens

				// Calculate cost based on model pricing (simplified)
				cost := (float64(requestTokens)/1000)*models[modelIdx].InputPrice + (float64(responseTokens)/1000)*models[modelIdx].OutputPrice

				billing := &model.BillingRecord{
					UserID:         user.ID,
					APIKeyID:       &apiKeyID,
					ModelID:        models[modelIdx].ID,
					RequestTokens:  requestTokens,
					ResponseTokens: responseTokens,
					TotalTokens:    totalTokens,
					Cost:           cost,
					Metadata:       model.JSONB{"path": "/v1/chat/completions", "status": "success"},
					CreatedAt:      time.Now().Add(-time.Duration(j*6) * time.Hour),
				}

				if err := db.WithContext(ctx).Create(billing).Error; err != nil {
					log.Printf("Failed to create billing record: %v", err)
				} else {
					fmt.Printf("✅ Created billing record for user %s with model %s\n", user.Email, models[modelIdx].Name)
				}
			}
		}

		fmt.Println("\nSeeding model statistics...")
		// Create statistics for last 7 days for each model
		for _, m := range models {
			for daysAgo := 0; daysAgo < 7; daysAgo++ {
				date := time.Now().Add(-time.Duration(daysAgo*24) * time.Hour)
				stat := &model.ModelStatistic{
					ModelID:         m.ID,
					Date:            date,
					TotalRequests:   1000 + int(m.ID[0])%1000 - daysAgo*100,
					TotalTokens:     50000 + int(m.ID[1])%50000 - daysAgo*5000,
					AvgResponseTime: 1.5 + float64(daysAgo%3)*0.2,
					SuccessRate:     95.0 + float64(daysAgo%5),
				}

				var existing model.ModelStatistic
				if err := db.WithContext(ctx).Where("model_id = ? AND date = ?", m.ID, date.Format("2006-01-02")).First(&existing).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						if err := db.WithContext(ctx).Create(stat).Error; err != nil {
							log.Printf("Failed to create model statistic: %v", err)
						} else {
							fmt.Printf("✅ Created statistic for model %s on %s\n", m.Name, date.Format("2006-01-02"))
						}
					}
				} else {
					fmt.Printf("⚠️ Model statistic exists for %s on %s\n", m.Name, date.Format("2006-01-02"))
				}
			}
		}
	}

	fmt.Println("\n✅ Seed data completed!")
}
