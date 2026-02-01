package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
	googleOAuth "golang.org/x/oauth2/google"
	"massrouter.ai/backend/internal/model"
	"massrouter.ai/backend/internal/repository"
	"massrouter.ai/backend/pkg/auth"
	"massrouter.ai/backend/pkg/utils"
)

type oauthService struct {
	userRepo          repository.UserRepository
	oauthProviderRepo repository.OAuthProviderRepository
	oauthAccountRepo  repository.OAuthAccountRepository
	jwtManager        *auth.JWTManager
	validator         *validator.Validate
}

func NewOAuthService(
	userRepo repository.UserRepository,
	oauthProviderRepo repository.OAuthProviderRepository,
	oauthAccountRepo repository.OAuthAccountRepository,
	jwtManager *auth.JWTManager,
) OAuthService {
	return &oauthService{
		userRepo:          userRepo,
		oauthProviderRepo: oauthProviderRepo,
		oauthAccountRepo:  oauthAccountRepo,
		jwtManager:        jwtManager,
		validator:         validator.New(),
	}
}

func (s *oauthService) GetEnabledProviders(ctx context.Context) ([]*OAuthProviderInfo, error) {
	providers, err := s.oauthProviderRepo.FindEnabledProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled providers: %w", err)
	}

	var result []*OAuthProviderInfo
	for _, provider := range providers {
		config := make(map[string]interface{})
		if provider.Config != nil {
			config = provider.Config
		}

		info := &OAuthProviderInfo{
			Name:        provider.Name,
			DisplayName: getDisplayName(provider.Config),
			ClientID:    provider.ClientID,
			AuthURL:     getAuthURL(provider.Config),
			TokenURL:    getTokenURL(provider.Config),
			UserInfoURL: getUserInfoURL(provider.Config),
			Scopes:      getScopes(provider.Config),
			Enabled:     provider.Enabled,
			Config:      config,
		}
		result = append(result, info)
	}

	return result, nil
}

func (s *oauthService) StartOAuthFlow(ctx context.Context, req *StartOAuthFlowRequest) (string, error) {
	if err := s.validator.Struct(req); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	provider, err := s.oauthProviderRepo.FindByName(ctx, req.Provider)
	if err != nil {
		return "", fmt.Errorf("provider not found: %w", err)
	}

	if !provider.Enabled {
		return "", fmt.Errorf("provider disabled")
	}

	config := make(map[string]interface{})
	if provider.Config != nil {
		config = provider.Config
	}

	oauthConfig := &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		RedirectURL:  req.CallbackURL,
		Scopes:       strings.Split(getScopes(config), " "),
	}

	// Use appropriate endpoints based on provider
	switch provider.Name {
	case "github":
		oauthConfig.Endpoint = githubOAuth.Endpoint
	case "google":
		oauthConfig.Endpoint = googleOAuth.Endpoint
	default:
		oauthConfig.Endpoint = oauth2.Endpoint{
			AuthURL:  getAuthURL(config),
			TokenURL: getTokenURL(config),
		}
	}

	state := req.State
	if state == "" {
		state = generateState()
	}

	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return authURL, nil
}

func (s *oauthService) HandleOAuthCallback(ctx context.Context, req *HandleOAuthCallbackRequest) (*OAuthLoginResponse, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	provider, err := s.oauthProviderRepo.FindByName(ctx, req.Provider)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	if !provider.Enabled {
		return nil, fmt.Errorf("provider disabled")
	}

	// Exchange code for token
	token, err := s.exchangeCodeForToken(ctx, provider, req.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from provider
	userInfo, err := s.getUserInfo(ctx, provider, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, isNewUser, err := s.findOrCreateUser(ctx, provider, userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	// Create or update OAuth account
	err = s.createOrUpdateOAuthAccount(ctx, user.ID, provider.ID, userInfo, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth account: %w", err)
	}

	// Update last login time
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	// Generate JWT tokens
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &OAuthLoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		IsNewUser:    isNewUser,
	}, nil
}

func (s *oauthService) DisconnectOAuthAccount(ctx context.Context, userID string, req *DisconnectOAuthAccountRequest) error {
	if err := s.validator.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	provider, err := s.oauthProviderRepo.FindByName(ctx, req.Provider)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	// Find the OAuth account
	account, err := s.oauthAccountRepo.FindByProviderAndUserID(ctx, provider.ID, userID)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	// Check if this is the last authentication method
	user, err := s.userRepo.FindWithOAuthAccounts(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user accounts: %w", err)
	}

	// Count authentication methods
	authMethods := 0
	if user.PasswordHash != "" {
		authMethods++
	}
	if len(user.OAuthAccounts) > 0 {
		authMethods += len(user.OAuthAccounts)
	}

	if authMethods <= 1 {
		return fmt.Errorf("cannot disconnect last authentication method")
	}

	// Delete the OAuth account
	if err := s.oauthAccountRepo.Delete(ctx, account.ID); err != nil {
		return fmt.Errorf("failed to delete OAuth account: %w", err)
	}

	return nil
}

func (s *oauthService) exchangeCodeForToken(ctx context.Context, provider *model.OAuthProvider, code string) (*oauth2.Token, error) {
	config := make(map[string]interface{})
	if provider.Config != nil {
		config = provider.Config
	}

	oauthConfig := &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		Scopes:       strings.Split(getScopes(config), " "),
	}

	switch provider.Name {
	case "github":
		oauthConfig.Endpoint = githubOAuth.Endpoint
	case "google":
		oauthConfig.Endpoint = googleOAuth.Endpoint
	default:
		oauthConfig.Endpoint = oauth2.Endpoint{
			AuthURL:  getAuthURL(config),
			TokenURL: getTokenURL(config),
		}
	}

	// Use http client with timeout
	httpClient := &http.Client{Timeout: 30 * time.Second}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	return token, nil
}

func (s *oauthService) getUserInfo(ctx context.Context, provider *model.OAuthProvider, accessToken string) (map[string]interface{}, error) {
	config := make(map[string]interface{})
	if provider.Config != nil {
		config = provider.Config
	}

	userInfoURL := getUserInfoURL(config)
	if userInfoURL == "" {
		return nil, fmt.Errorf("user info URL not configured for provider")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return userInfo, nil
}

func (s *oauthService) findOrCreateUser(ctx context.Context, provider *model.OAuthProvider, userInfo map[string]interface{}) (*model.User, bool, error) {
	// Extract email from user info
	email, ok := userInfo["email"].(string)
	if !ok || email == "" {
		// Try alternative email fields
		if emailAlt, ok := userInfo["email_address"].(string); ok && emailAlt != "" {
			email = emailAlt
		} else if emailAlt, ok := userInfo["user_email"].(string); ok && emailAlt != "" {
			email = emailAlt
		} else {
			return nil, false, fmt.Errorf("email not provided by provider")
		}
	}

	// Try to find existing user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && user != nil {
		return user, false, nil
	}

	// Create new user
	username := generateUsername(email, userInfo)

	// Generate random password for OAuth users (won't be used for login)
	passwordHash, err := utils.HashPassword(generateRandomPassword())
	if err != nil {
		return nil, false, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &model.User{
		Email:         email,
		Username:      username,
		PasswordHash:  passwordHash,
		Role:          "user",
		Status:        "active",
		EmailVerified: true, // OAuth providers typically verify emails
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, false, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser, true, nil
}

func (s *oauthService) createOrUpdateOAuthAccount(ctx context.Context, userID, providerID string, userInfo map[string]interface{}, token *oauth2.Token) error {
	providerUserID, ok := userInfo["id"].(string)
	if !ok {
		// Try to convert numeric ID to string
		if id, ok := userInfo["id"].(float64); ok {
			providerUserID = fmt.Sprintf("%.0f", id)
		} else {
			return fmt.Errorf("provider user ID not found in user info")
		}
	}

	// Try to find existing account
	account, err := s.oauthAccountRepo.FindByProviderUserID(ctx, providerID, providerUserID)
	if err == nil && account != nil {
		// Update existing account with new tokens
		var expiresAt *time.Time
		if !token.Expiry.IsZero() {
			expiresAt = &token.Expiry
		}
		return s.oauthAccountRepo.UpdateTokens(ctx, account.ID, token.AccessToken, token.RefreshToken, expiresAt)
	}

	// Create new account
	newAccount := &model.OAuthAccount{
		UserID:         userID,
		ProviderID:     providerID,
		ProviderUserID: providerUserID,
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
	}

	if !token.Expiry.IsZero() {
		newAccount.TokenExpiresAt = &token.Expiry
	}

	return s.oauthAccountRepo.Create(ctx, newAccount)
}

// Helper functions
func getDisplayName(config map[string]interface{}) string {
	if name, ok := config["display_name"].(string); ok {
		return name
	}
	return ""
}

func getAuthURL(config map[string]interface{}) string {
	if url, ok := config["auth_url"].(string); ok {
		return url
	}
	return ""
}

func getTokenURL(config map[string]interface{}) string {
	if url, ok := config["token_url"].(string); ok {
		return url
	}
	return ""
}

func getUserInfoURL(config map[string]interface{}) string {
	if url, ok := config["user_info_url"].(string); ok {
		return url
	}
	return ""
}

func getScopes(config map[string]interface{}) string {
	if scopes, ok := config["scopes"].(string); ok {
		return scopes
	}
	return ""
}

func generateState() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 32
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func generateUsername(email string, userInfo map[string]interface{}) string {
	// Try to get username from user info
	if username, ok := userInfo["login"].(string); ok && username != "" {
		return username
	}
	if username, ok := userInfo["username"].(string); ok && username != "" {
		return username
	}
	if username, ok := userInfo["name"].(string); ok && username != "" {
		// Convert name to username (lowercase, replace spaces)
		username = strings.ToLower(username)
		username = strings.ReplaceAll(username, " ", "_")
		return username
	}

	// Fallback to email prefix
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}

	// Final fallback
	return fmt.Sprintf("user_%d", time.Now().Unix())
}

func generateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	const length = 32
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
