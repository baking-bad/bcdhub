package oauth

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/cmd/api/providers"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// Config -
type Config struct {
	Providers      map[string]providers.Oauth
	JWTKey         []byte
	State          string
	UserID         uint
	JWTRedirectURL string
}

// New -
func New(cfg config.Config) (Config, error) {
	return Config{
		Providers:      providers.InitOauth(cfg),
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

	if time.Now().Unix() > claims.StandardClaims.ExpiresAt {
		return 0, fmt.Errorf("token expired")
	}

	return claims.UserID, nil
}
