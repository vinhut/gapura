package main

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mocks "github.com/vinhut/gapura/mocks"
	"github.com/vinhut/gapura/models"

	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestPing(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_user := mocks.NewMockUserDatabase(ctrl)

	router := setupRouter(mock_user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestLogin(t *testing.T) {

	test_email := "newusertest@test.com"
	test_password := "test_password"
	hashed_password := "$2a$14$quU8rC8Cfska91KtagAkhOPdMvQ5sEPMwelBFDFvrdUR2/uCoa/MC"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_user := mocks.NewMockUserDatabase(ctrl)
	input := models.User{}
	mock_user.EXPECT().Find("email", test_email, &input).Do(
		func(a string, b string, c *models.User) error {
			c.Email = test_email
			c.Password = hashed_password
			c.Role = "standard"
			return nil
		})

	router := setupRouter(mock_user)

	var param = url.Values{}
	param.Set("email", test_email)
	param.Set("password", test_password)
	var payload = bytes.NewBufferString(param.Encode())

	pre_w := httptest.NewRecorder()
	pre_req, _ := http.NewRequest("POST", "/login", payload)
	pre_req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(pre_w, pre_req)

	assert.Equal(t, 200, pre_w.Code)
}

func TestDecryptToken(t *testing.T) {

	service_name := "test-service"
	token := "852a37a34b727c0e0b331806-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b3c7fd784d4"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_user := mocks.NewMockUserDatabase(ctrl)

	router := setupRouter(mock_user)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user?service="+service_name+
		"&token="+token, nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
}

func TestPublicEndpoint(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_user := mocks.NewMockUserDatabase(ctrl)

	router := setupRouter(mock_user)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)

}
