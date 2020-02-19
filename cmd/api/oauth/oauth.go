package oauth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// Config -
type Config struct {
	Oauth2 *oauth2.Config
	JWTKey []byte
	State  string
	UserID uint
}

// New -
func New() (Config, error) {
	// TO-DO: uncomment in prod
	// var githubClientID, githubClientSecret, jwtKey, oauthStateString string

	// if id := os.Getenv("OAUTH_CLIENT_ID"); id == "" {
	// 	return nil, fmt.Errorf("emtpty OAUTH_CLIENT_ID env variable")
	// } else {
	// 	githubClientID = id
	// }

	// if secret := os.Getenv("OAUTH_CLIENT_SECRET"); secret == "" {
	// 	return nil, fmt.Errorf("emtpty OAUTH_CLIENT_SECRET env variable")
	// } else {
	// 	githubClientSecret = secret
	// }

	// if jwt := os.Getenv("JWT_SECRET_KEY"); jwt == "" {
	// 	return nil, fmt.Errorf("emtpty JWT_SECRET_KEY env variable")
	// } else {
	// 	jwtKey = jwt
	// }

	// if state := os.Getenv("OAUTH_STATE_STRING"); state == "" {
	// 	return nil, fmt.Errorf("emtpty OAUTH_STATE_STRING env variable")
	// } else {
	// 	oauthStateString = state
	// }

	// TO-DO: delete in prod
	githubClientID := "d35966939d838f410dd9"
	githubClientSecret := "287ae6a529f479afadd19e4e2386b33f5889f58c"
	jwtKey := []byte("my_secret_key")
	oauthStateString := "pseudo-random"

	return Config{
		Oauth2: &oauth2.Config{
			RedirectURL:  "http://localhost:14000/v1/oauth/callback",
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			Scopes:       []string{},
			Endpoint:     github.Endpoint,
		},
		JWTKey: jwtKey,
		State:  oauthStateString,
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
