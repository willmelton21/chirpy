package auth

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	
	claims := &jwt.RegisteredClaims{}


	token, err := jwt.ParseWithClaims(tokenString,claims, func(token *jwt.Token) (interface{}, error) {
		
		if _,ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(tokenSecret),nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return  userID, nil
}



func TestMakeAndValidateJWT(t *testing.T) {

	userID := uuid.New()
	secret := "test-secret"

	//Test 1: Create and avlidate a toekn
	token, err := MakeJWT(userID, secret, time.Hour)
	assert.NoError(t,err)
	assert.NotEmpty(t,token)

	//Validate the token
	extractedID, err := ValidateJWT(token,secret)
	assert.NoError(t,err)
	assert.Equal(t, userID, extractedID)

	//Test 2: Expired token
	expiredToken, err := MakeJWT(userID,secret, -time.Hour)
	assert.NoError(t,err)

	_,err = ValidateJWT(expiredToken,secret)
	assert.Error(t,err)

	wrongSecret := "wrong-secret"
	_, err = ValidateJWT(token, wrongSecret)
	assert.Error(t, err)

	_, err = ValidateJWT("not.a.valid.token",secret)
	assert.Error(t,err)


}
