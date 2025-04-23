package auth

import (

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
