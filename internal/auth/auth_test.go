package auth

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/google/uuid"
	"time"
	"net/http"
)





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

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name: "Valid Bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer valid_token"},
			},
			wantToken: "valid_token",
			wantErr:   false,
		},
		{
			name:      "Missing Authorization header",
			headers:   http.Header{},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "Malformed Authorization header",
			headers: http.Header{
				"Authorization": []string{"InvalidBearer token"},
			},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotToken != tt.wantToken {
				t.Errorf("GetBearerToken() gotToken = %v, want %v", gotToken, tt.wantToken)
			}
		})
	}
}
