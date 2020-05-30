package main

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mocks "github.com/vinhut/gapura/mocks"

	"net/http"
	"net/http/httptest"
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
