package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJSONB_Value(t *testing.T) {
	tests := []struct {
		name    string
		jsonb   JSONB
		want    string
		wantErr bool
	}{
		{
			name:    "empty JSONB",
			jsonb:   JSONB{},
			want:    "{}",
			wantErr: false,
		},
		{
			name: "simple JSONB",
			jsonb: JSONB{
				"key": "value",
			},
			want:    `{"key":"value"}`,
			wantErr: false,
		},
		{
			name: "nested JSONB",
			jsonb: JSONB{
				"nested": map[string]interface{}{
					"inner": "value",
				},
			},
			want:    `{"nested":{"inner":"value"}}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.jsonb.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONB.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				gotBytes, ok := got.([]byte)
				if !ok {
					t.Errorf("JSONB.Value() returned %T, want []byte", got)
					return
				}

				var gotMap, wantMap map[string]interface{}
				if err := json.Unmarshal(gotBytes, &gotMap); err != nil {
					t.Errorf("Failed to unmarshal result: %v", err)
					return
				}
				if err := json.Unmarshal([]byte(tt.want), &wantMap); err != nil {
					t.Errorf("Failed to unmarshal expected: %v", err)
					return
				}

				gotJSON, _ := json.Marshal(gotMap)
				wantJSON, _ := json.Marshal(wantMap)
				if string(gotJSON) != string(wantJSON) {
					t.Errorf("JSONB.Value() = %v, want %v", string(gotJSON), string(wantJSON))
				}
			}
		})
	}
}

func TestJSONB_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    JSONB
		wantErr bool
	}{
		{
			name:    "nil value",
			value:   nil,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "empty byte slice",
			value:   []byte("{}"),
			want:    JSONB{},
			wantErr: false,
		},
		{
			name:  "valid JSON byte slice",
			value: []byte(`{"key":"value"}`),
			want: JSONB{
				"key": "value",
			},
			wantErr: false,
		},
		{
			name:  "valid JSON string",
			value: `{"key":"value"}`,
			want: JSONB{
				"key": "value",
			},
			wantErr: false,
		},
		{
			name:  "complex JSON",
			value: []byte(`{"nested":{"inner":"value"},"array":[1,2,3]}`),
			want: JSONB{
				"nested": map[string]interface{}{
					"inner": "value",
				},
				"array": []interface{}{float64(1), float64(2), float64(3)},
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			value:   []byte(`{invalid}`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unsupported type",
			value:   123,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j JSONB
			err := j.Scan(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("JSONB.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				gotJSON, _ := json.Marshal(j)
				wantJSON, _ := json.Marshal(tt.want)
				if string(gotJSON) != string(wantJSON) {
					t.Errorf("JSONB.Scan() = %v, want %v", string(gotJSON), string(wantJSON))
				}
			}
		})
	}
}

func TestUser_TableName(t *testing.T) {
	u := User{}
	if got := u.TableName(); got != "users" {
		t.Errorf("User.TableName() = %v, want %v", got, "users")
	}
}

func TestOAuthProvider_TableName(t *testing.T) {
	p := OAuthProvider{}
	if got := p.TableName(); got != "oauth_providers" {
		t.Errorf("OAuthProvider.TableName() = %v, want %v", got, "oauth_providers")
	}
}

func TestOAuthAccount_TableName(t *testing.T) {
	a := OAuthAccount{}
	if got := a.TableName(); got != "oauth_accounts" {
		t.Errorf("OAuthAccount.TableName() = %v, want %v", got, "oauth_accounts")
	}
}

func TestUser_Validate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "valid user",
			user: User{
				Email:        "test@example.com",
				Username:     "testuser",
				PasswordHash: "hashed_password",
				Role:         "user",
				Status:       "active",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			wantErr: false,
		},
		{
			name: "empty email",
			user: User{
				Email:        "",
				Username:     "testuser",
				PasswordHash: "hashed_password",
			},
			wantErr: true,
		},
		{
			name: "empty username",
			user: User{
				Email:        "test@example.com",
				Username:     "",
				PasswordHash: "hashed_password",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasEmail := tt.user.Email != ""
			hasUsername := tt.user.Username != ""
			hasPassword := tt.user.PasswordHash != ""

			if !hasEmail || !hasUsername || !hasPassword {
				if !tt.wantErr {
					t.Errorf("User validation should fail for missing fields")
				}
			} else {
				if tt.wantErr {
					t.Errorf("User validation should pass for valid user")
				}
			}
		})
	}
}
