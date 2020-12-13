package main

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mocks "github.com/vinhut/gapura/mocks"
	"github.com/vinhut/gapura/models"
	"go.mongodb.org/mongo-driver/bson/primitive"

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

func TestFailedLogin(t *testing.T) {

	test_email := "newusertest@test.com"
	test_password := "WRONG Password"
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

	assert.Equal(t, 401, pre_w.Code)
}

func TestDecryptToken(t *testing.T) {

	service_name := "test-service"
	test_uid := primitive.NewObjectID()
	token := "439e835079c54999066756ce-826dc68c5fc8c80aafab394632223f64dabdf89f0fd1b1a08d2398fed435d8b89ed0d2a97748054a1f442cf2b6adb953bbd4d5318f7dd7304dda9562325bad2b6d29843121690e7d7e27e1cdd17b9fef9be49aff66cd7610698b7fdd4722528432b545a522fabc334c9fb3a2b07299d4ebdc3d70c1b0204ea88b38da670447b9031d2fe9235a56cb268d"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_user := mocks.NewMockUserDatabase(ctrl)
	input := models.User{}
	mock_user.EXPECT().FindByUid("_id", gomock.Any(), &input).Do(
		func(a string, b string, c *models.User) error {
			c.Uid = test_uid
			return nil
		})

	router := setupRouter(mock_user)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user?service="+service_name+
		"&token="+token, nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
}

func TestFailedDecryptToken(t *testing.T) {

	service_name := "test-service"
	token := "852a37a34b727c0e0b3318ab-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b3c7" //fd784d4"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_user := mocks.NewMockUserDatabase(ctrl)

	router := setupRouter(mock_user)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user?service="+service_name+
		"&token="+token, nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, 401, rec.Code)
}

func TestDecryptTokenFromCookies(t *testing.T) {

	service_name := "test-service"
	test_uid := primitive.NewObjectID()
	token := "439e835079c54999066756ce-826dc68c5fc8c80aafab394632223f64dabdf89f0fd1b1a08d2398fed435d8b89ed0d2a97748054a1f442cf2b6adb953bbd4d5318f7dd7304dda9562325bad2b6d29843121690e7d7e27e1cdd17b9fef9be49aff66cd7610698b7fdd4722528432b545a522fabc334c9fb3a2b07299d4ebdc3d70c1b0204ea88b38da670447b9031d2fe9235a56cb268d"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_user := mocks.NewMockUserDatabase(ctrl)
	input := models.User{}
	mock_user.EXPECT().FindByUid("_id", gomock.Any(), &input).Do(
		func(a string, b string, c *models.User) error {
			c.Uid = test_uid
			return nil
		})

	router := setupRouter(mock_user)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user?service="+service_name, nil)
	req.Header.Set("Cookie", "token="+token+";")
	router.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
}

func TestFailedDecryptTokenFromCookies(t *testing.T) {

	service_name := "test-service"
	token := "852a37a34b727c0e0b331806-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b" //3c7fd784d4"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_user := mocks.NewMockUserDatabase(ctrl)

	router := setupRouter(mock_user)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user?service="+service_name, nil)
	req.Header.Set("Cookie", "token="+token+";")
	router.ServeHTTP(rec, req)

	assert.Equal(t, 401, rec.Code)
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

func TestCreateUser(t *testing.T) {

	test_email := "newusertest@test.com"
	test_password := "test_password"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_user := mocks.NewMockUserDatabase(ctrl)
	mock_user.EXPECT().Create(gomock.Any()).Return(true, nil)

	router := setupRouter(mock_user)

	var param = url.Values{}
	param.Set("email", test_email)
	param.Set("password", test_password)
	var payload = bytes.NewBufferString(param.Encode())

	pre_w := httptest.NewRecorder()
	pre_req, _ := http.NewRequest("POST", "/user?service=test", payload)
	pre_req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(pre_w, pre_req)

	assert.Equal(t, 200, pre_w.Code)

}
