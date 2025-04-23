package auth

import (
	"time"
	
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	
)

func HashPassword(password string) (string, error) {

	encryptedPass, err := bcrypt.GenerateFromPassword([]byte(password),10)
	if err != nil {
		return "",err
	}

	return string(encryptedPass),nil

}

func CheckPasswordHash(hash,password string) error {

	err := bcrypt.CompareHashAndPassword([]byte(hash),[]byte(password))
	if err != nil {
		return err
	}
	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecrets string, expiresIn time.Duration) (string,error) {


	claims := jwt.RegisteredClaims{Issuer: "chirpy",
								IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
								ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
								Subject: userID.String()}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedStr, err := token.SignedString([]byte(tokenSecrets))
	if err != nil {
		return "",err
	}

	return signedStr,nil
}


