package oauth

import (
	"fmt"
	"os"
	"time"

	"github.com/baking-bad/bcdhub/internal/jsonload"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/gitlab"
)

// InitConfig -
type InitConfig struct {
	GithubCallbackURL string `json:"githubCallbackURL"`
	GitlabCallbackURL string `json:"gitlabCallbackURL"`
	JwtRedirectURL    string `json:"jwtRedirectURL"`
}

// Config -
type Config struct {
	Github         *oauth2.Config
	Gitlab         *oauth2.Config
	JWTKey         []byte
	State          string
	UserID         uint
	JWTRedirectURL string
}

// New -
func New() (Config, error) {
	envVars := []string{
		"GITHUB_CLIENT_ID",
		"GITHUB_CLIENT_SECRET",
		"GITLAB_CLIENT_ID",
		"GITLAB_CLIENT_SECRET",
		"JWT_SECRET_KEY",
		"OAUTH_STATE_STRING",
	}

	for _, ev := range envVars {
		if os.Getenv(ev) == "" {
			return Config{}, fmt.Errorf("empty %s env variable", ev)
		}
	}

	githubClientID := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	gitlabClientID := os.Getenv("GITLAB_CLIENT_ID")
	gitlabClientSecret := os.Getenv("GITLAB_CLIENT_SECRET")
	jwtKey := os.Getenv("JWT_SECRET_KEY")
	oauthStateString := os.Getenv("OAUTH_STATE_STRING")

	configName := "development.json"
	if env := os.Getenv("BCD_ENV"); env == "production" {
		configName = "production.json"
	}

	var cfg InitConfig
	if err := jsonload.StructFromFile(fmt.Sprintf("./oauth/%s", configName), &cfg); err != nil {
		return Config{}, err
	}

	return Config{
		Github: &oauth2.Config{
			RedirectURL:  cfg.GithubCallbackURL,
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			Scopes:       []string{},
			Endpoint:     github.Endpoint,
		},
		Gitlab: &oauth2.Config{
			RedirectURL:  cfg.GitlabCallbackURL,
			ClientID:     gitlabClientID,
			ClientSecret: gitlabClientSecret,
			Scopes:       []string{"read_user"},
			Endpoint:     gitlab.Endpoint,
		},
		JWTKey:         []byte(jwtKey),
		State:          oauthStateString,
		JWTRedirectURL: cfg.JwtRedirectURL,
	}, nil
}

type jwtClaims struct {
	UserID uint `json:"userID"`
	jwt.StandardClaims
}

// MakeJWT -
func (c Config) MakeJWT(userID uint) (string, error) {
	expirationTime := time.Now().Add(48 * time.Hour)

	claims := &jwtClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(c.JWTKey))
}

// GetIDFromToken -
func (c Config) GetIDFromToken(token string) (uint, error) {
	claims := &jwtClaims{}

	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.JWTKey), nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to parse token %v", token)
	}

	if !tkn.Valid {
		return 0, fmt.Errorf("invalid token %v", token)
	}

	return claims.UserID, nil
}
