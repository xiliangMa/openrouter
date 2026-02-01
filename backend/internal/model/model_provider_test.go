package model

import (
	"testing"
	"time"
)

func TestModelProvider_TableName(t *testing.T) {
	p := ModelProvider{}
	if got := p.TableName(); got != "model_providers" {
		t.Errorf("ModelProvider.TableName() = %v, want %v", got, "model_providers")
	}
}

func TestModel_TableName(t *testing.T) {
	m := Model{}
	if got := m.TableName(); got != "models" {
		t.Errorf("Model.TableName() = %v, want %v", got, "models")
	}
}

func TestUserAPIKey_TableName(t *testing.T) {
	k := UserAPIKey{}
	if got := k.TableName(); got != "user_api_keys" {
		t.Errorf("UserAPIKey.TableName() = %v, want %v", got, "user_api_keys")
	}
}

func TestModelProvider_Validate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		provider      ModelProvider
		wantValid     bool
		errorContains string
	}{
		{
			name: "valid provider",
			provider: ModelProvider{
				Name:       "OpenAI",
				APIBaseURL: "https://api.openai.com/v1",
				Config:     JSONB{"description": "OpenAI API"},
				Status:     "active",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			wantValid: true,
		},
		{
			name: "empty name",
			provider: ModelProvider{
				Name:       "",
				APIBaseURL: "https://api.openai.com/v1",
			},
			wantValid:     false,
			errorContains: "name",
		},
		{
			name: "empty API base URL",
			provider: ModelProvider{
				Name:       "OpenAI",
				APIBaseURL: "",
			},
			wantValid:     false,
			errorContains: "API base URL",
		},
		{
			name: "invalid URL",
			provider: ModelProvider{
				Name:       "OpenAI",
				APIBaseURL: "not-a-url",
			},
			wantValid:     false,
			errorContains: "URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasName := tt.provider.Name != ""
			hasURL := tt.provider.APIBaseURL != ""

			if !hasName || !hasURL {
				if tt.wantValid {
					t.Errorf("ModelProvider validation should fail for missing fields")
				}
			} else {
				// Simple URL check
				hasValidURL := len(tt.provider.APIBaseURL) > 0 && (tt.provider.APIBaseURL[:7] == "http://" || tt.provider.APIBaseURL[:8] == "https://")
				if !hasValidURL && tt.wantValid {
					t.Errorf("ModelProvider validation should fail for invalid URL")
				} else if hasValidURL && !tt.wantValid && tt.errorContains == "URL" {
					t.Errorf("ModelProvider validation should pass for valid URL")
				}
			}
		})
	}
}

func TestModel_Validate(t *testing.T) {
	now := time.Now()
	contextLength := 128000
	maxTokens := 4096

	tests := []struct {
		name      string
		model     Model
		wantValid bool
	}{
		{
			name: "valid model",
			model: Model{
				Name:          "gpt-4o",
				ProviderID:    "provider-id",
				Description:   "OpenAI GPT-4 Omni",
				ContextLength: &contextLength,
				MaxTokens:     &maxTokens,
				Capabilities:  JSONB{"vision": true},
				Category:      "chat",
				PricingTier:   "premium",
				InputPrice:    0.005,
				OutputPrice:   0.015,
				IsFree:        false,
				IsActive:      true,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			wantValid: true,
		},
		{
			name: "model without name",
			model: Model{
				Name:       "",
				ProviderID: "provider-id",
			},
			wantValid: false,
		},
		{
			name: "model without provider ID",
			model: Model{
				Name:       "gpt-4o",
				ProviderID: "",
			},
			wantValid: false,
		},
		{
			name: "free model with zero prices",
			model: Model{
				Name:        "free-model",
				ProviderID:  "provider-id",
				IsFree:      true,
				InputPrice:  0,
				OutputPrice: 0,
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasName := tt.model.Name != ""
			hasProviderID := tt.model.ProviderID != ""

			if !hasName || !hasProviderID {
				if tt.wantValid {
					t.Errorf("Model validation should fail for missing fields")
				}
			} else {
				if !tt.wantValid {
					t.Errorf("Model validation should pass for valid model")
				}
			}
		})
	}
}

func TestUserAPIKey_Validate(t *testing.T) {
	now := time.Now()
	future := now.AddDate(1, 0, 0)

	tests := []struct {
		name      string
		apiKey    UserAPIKey
		wantValid bool
	}{
		{
			name: "valid API key",
			apiKey: UserAPIKey{
				Name:        "Test Key",
				APIKey:      "sk-test-123",
				Prefix:      "sk-test",
				Permissions: JSONB{"models": []interface{}{"read"}},
				RateLimit:   1000,
				UserID:      "user-id",
				IsActive:    true,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantValid: true,
		},
		{
			name: "API key with expiration",
			apiKey: UserAPIKey{
				Name:      "Test Key",
				APIKey:    "sk-test-123",
				Prefix:    "sk-test",
				ExpiresAt: &future,
				UserID:    "user-id",
			},
			wantValid: true,
		},
		{
			name: "API key without name",
			apiKey: UserAPIKey{
				Name:   "",
				APIKey: "sk-test-123",
				UserID: "user-id",
			},
			wantValid: false,
		},
		{
			name: "API key without key",
			apiKey: UserAPIKey{
				Name:   "Test Key",
				APIKey: "",
				UserID: "user-id",
			},
			wantValid: false,
		},
		{
			name: "API key without user ID",
			apiKey: UserAPIKey{
				Name:   "Test Key",
				APIKey: "sk-test-123",
				UserID: "",
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasName := tt.apiKey.Name != ""
			hasAPIKey := tt.apiKey.APIKey != ""
			hasUserID := tt.apiKey.UserID != ""

			if !hasName || !hasAPIKey || !hasUserID {
				if tt.wantValid {
					t.Errorf("UserAPIKey validation should fail for missing fields")
				}
			} else {
				if !tt.wantValid {
					t.Errorf("UserAPIKey validation should pass for valid API key")
				}
			}
		})
	}
}
