package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenType string

const (
	// TokenTypeAccess -
	TokenTypeAccess TokenType = "chirpy-access"
)

// ErrNoAuthHeaderIncluded -
var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	encodedKey := hex.EncodeToString(key)
	return encodedKey, nil
}

func GetAPIKey(headers http.Header) (string, error) {

	authHeader := headers.Get("Authorization")	
	if authHeader == ""{
		return "",fmt.Errorf("Couldn't Get Header Auth")
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		return "",fmt.Errorf("Malformed Auth Header")
	}

	token := splitAuth[1]

	return token,nil
}

// HashPassword -
func HashPassword(password string) (string, error) {
	dat, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

// CheckPasswordHash -
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GetUserIDFromToken(tokenString string) (uuid.UUID, error) {
 // Parse the token
	tokenSecret := os.Getenv("SECRET")
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(tokenSecret), nil
    })
    
    if err != nil {
        return uuid.Nil, err
    }
    
    // Extract claims
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        // Get the subject claim which contains the user ID
        sub, ok := claims["sub"].(string)
        if !ok {
            return uuid.Nil, fmt.Errorf("invalid subject in token")
        }
        
        // Parse the string back to UUID
        userID, err := uuid.Parse(sub)
        if err != nil {
            return uuid.Nil, fmt.Errorf("invalid user ID format in token")
        }
        
        return userID, nil
    }
    
    return uuid.Nil, fmt.Errorf("invalid token")


}

// MakeJWT -
func MakeJWT(
	userID uuid.UUID,
	tokenSecret string,
	expiresIn time.Duration,
) (string, error) {
	signingKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString(signingKey)
}

// ValidateJWT -
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.Nil, err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	id, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return id, nil
}

// GetBearerToken -
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}

	return splitAuth[1], nil
}
