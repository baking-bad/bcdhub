package oauth

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
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
func New(cfg config.Config) (Config, error) {
	return Config{
		Github: &oauth2.Config{
			RedirectURL:  cfg.OAuth.Github.CallbackURL,
			ClientID:     cfg.OAuth.Github.ClientID,
			ClientSecret: cfg.OAuth.Github.Secret,
			Scopes:       []string{},
			Endpoint:     github.Endpoint,
		},
		Gitlab: &oauth2.Config{
			RedirectURL:  cfg.OAuth.Gitlab.CallbackURL,
			ClientID:     cfg.OAuth.Gitlab.ClientID,
			ClientSecret: cfg.OAuth.Gitlab.Secret,
			Scopes:       []string{"read_user"},
			Endpoint:     gitlab.Endpoint,
		},
		JWTKey:         []byte(cfg.OAuth.JWT.Secret),
		State:          cfg.OAuth.State,
		JWTRedirectURL: cfg.OAuth.JWT.RedirectURL,
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
		return 0, errors.Errorf("failed to parse token %v", token)
	}

	if !tkn.Valid {
		return 0, errors.Errorf("invalid token %v", token)
	}

	return claims.UserID, nil
}
