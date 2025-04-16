package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"gorm.io/gorm"
)

// OAuthService manages OAuth authentication
type OAuthService struct {
	githubConfig *oauth2.Config
}

// NewOAuthService creates a new OAuth service with Github and Google configs
func NewOAuthService() *OAuthService {
	// GitHub OAuth config
	githubConfig := &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_BASE") + "/auth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	return &OAuthService{
		githubConfig: githubConfig,
	}
}

// GenerateState generates a random state string for OAuth
func (s *OAuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// GetGithubAuthURL returns the GitHub authorization URL
func (s *OAuthService) GetGithubAuthURL(state string) string {
	return s.githubConfig.AuthCodeURL(state)
}

// GithubUser represents GitHub user information
type GithubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// HandleGithubCallback processes GitHub OAuth callback
func (s *OAuthService) HandleGithubCallback(code string) (*models.User, error) {
	// Exchange code for token
	token, err := s.githubConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("GitHub code exchange failed: %v", err)
	}

	// Get user info
	client := s.githubConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("Failed to get user info from GitHub: %v", err)
	}
	defer resp.Body.Close()

	// Parse user data
	var githubUser GithubUser
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read GitHub response: %v", err)
	}
	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, fmt.Errorf("Failed to parse GitHub user data: %v", err)
	}

	// If email is not provided, get it from emails endpoint
	if githubUser.Email == "" {
		emailsResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailsResp.Body.Close()
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}

			emailsBody, _ := io.ReadAll(emailsResp.Body)
			json.Unmarshal(emailsBody, &emails)

			// Find primary email
			for _, email := range emails {
				if email.Primary && email.Verified {
					githubUser.Email = email.Email
					break
				}
			}
		}
	}

	// Use login as username if name is empty
	username := githubUser.Name
	if username == "" {
		username = githubUser.Login
	}

	return s.findOrCreateUser(githubUser.Email, username, "github")
}

// findOrCreateUser finds existing user or creates a new one
func (s *OAuthService) findOrCreateUser(email, username, provider string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("Email not provided by %s", provider)
	}

	var user models.User
	result := database.DB.Where("email = ?", email).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create a new user with a random password
			randomBytes := make([]byte, 32)
			_, err := rand.Read(randomBytes)
			if err != nil {
				return nil, fmt.Errorf("Failed to generate random password: %v", err)
			}
			randomPassword := base64.StdEncoding.EncodeToString(randomBytes)

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
			if err != nil {
				return nil, fmt.Errorf("Failed to hash password: %v", err)
			}

			newUser := models.User{
				Email:     email,
				Username:  username,
				Password:  string(hashedPassword),
				CreatedAt: time.Now(),
			}

			if err := database.DB.Create(&newUser).Error; err != nil {
				return nil, fmt.Errorf("Failed to create user: %v", err)
			}

			return &newUser, nil
		}
		return nil, fmt.Errorf("Database error: %v", result.Error)
	}

	// Update username if it's empty
	if user.Username == "" && username != "" {
		user.Username = username
		database.DB.Save(&user)
	}

	return &user, nil
}
