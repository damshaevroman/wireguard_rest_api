package controllers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"wireguard_api/config"

	"wireguard_api/usecases"

	"github.com/gin-gonic/gin"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func setupGin(method, path string, h gin.HandlerFunc) (*gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Handle(method, path, h)
	return r, httptest.NewRecorder()
}

func TestGetStatus_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUsecaseService(ctrl)

	mockSvc.EXPECT().
		GetStatus().
		Return([]usecases.InterfaceListStatus{
			{Ifname: "wg0", Status: []usecases.Status{}},
		}, nil)

	controller := NewController(mockSvc, &config.ServerConfig{})
	r, w := setupGin("GET", "/status", controller.GetStatus)

	req, _ := http.NewRequest("GET", "/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetStatus_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUsecaseService(ctrl)

	mockSvc.EXPECT().
		GetStatus().
		Return(nil, errors.New("service error"))

	controller := NewController(mockSvc, &config.ServerConfig{})
	r, w := setupGin("GET", "/status", controller.GetStatus)

	req, _ := http.NewRequest("GET", "/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetAllClients_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUsecaseService(ctrl)

	mockSvc.EXPECT().
		GetAllClients().
		Return([]usecases.ClientResponse{
			{},
		}, nil)

	controller := NewController(mockSvc, &config.ServerConfig{})
	r, w := setupGin("GET", "/clients", controller.GetAllClients)

	req, _ := http.NewRequest("GET", "/clients", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetAllClients_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUsecaseService(ctrl)

	mockSvc.EXPECT().
		GetAllClients().
		Return(nil, errors.New("error"))

	controller := NewController(mockSvc, &config.ServerConfig{})
	r, w := setupGin("GET", "/clients", controller.GetAllClients)

	req, _ := http.NewRequest("GET", "/clients", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteClient_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUsecaseService(ctrl)

	mockSvc.EXPECT().
		DeleteClient("pubkey").
		Return(nil)

	cfg := &config.ServerConfig{ClientDelete: true}
	controller := NewController(mockSvc, cfg)

	body := `{"public":"pubkey"}`
	r, w := setupGin("DELETE", "/client", controller.DeleteClient)

	req, _ := http.NewRequest("DELETE", "/client", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteClient_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUsecaseService(ctrl)

	cfg := &config.ServerConfig{ClientDelete: false}
	controller := NewController(mockSvc, cfg)

	r, w := setupGin("DELETE", "/client", controller.DeleteClient)

	req, _ := http.NewRequest("DELETE", "/client", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
func TestAddClient_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUsecaseService(ctrl)

	mockSvc.EXPECT().
		NewClient("wg0", "10.0.0.2/32", "0.0.0.0/0").
		Return(usecases.ClientResponse{Ifname: "wg0"}, nil)

	controller := NewController(mockSvc, &config.ServerConfig{})

	body := `{
		"ifname":"wg0",
		"ip":"10.0.0.2/32",
		"alloweip":"0.0.0.0/0"
	}`

	r, w := setupGin("POST", "/client", controller.AddClient)

	req, _ := http.NewRequest("POST", "/client", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddClient_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUsecaseService(ctrl)

	mockSvc.EXPECT().
		NewClient(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(usecases.ClientResponse{}, errors.New("create error"))

	controller := NewController(mockSvc, &config.ServerConfig{})

	body := `{
		"ifname":"wg0",
		"ip":"10.0.0.2/32",
		"allowed_ip":"0.0.0.0/0"
	}`

	r, w := setupGin("POST", "/client", controller.AddClient)

	req, _ := http.NewRequest("POST", "/client", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAddInterface_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUsecaseService(ctrl)

	mockSvc.EXPECT().
		NewInterface("wg0", "10.0.0.1/24", "1.2.3.4", 51820).
		Return(usecases.ServerInterfaces{Ifname: "wg0"}, nil)

	controller := NewController(mockSvc, &config.ServerConfig{})

	body := `{
		"ifname":"wg0",
		"ip":"10.0.0.1/24",
		"endpoint":"1.2.3.4",
		"port":51820
	}`

	r, w := setupGin("POST", "/interface", controller.AddInterface)

	req, _ := http.NewRequest("POST", "/interface", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddInterface_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUsecaseService(ctrl)

	mockSvc.EXPECT().
		NewInterface(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(usecases.ServerInterfaces{}, errors.New("fail"))

	controller := NewController(mockSvc, &config.ServerConfig{})

	body := `{
		"ifname":"wg0",
		"ip":"10.0.0.1/24",
		"endpoint":"1.2.3.4",
		"port":51820
	}`

	r, w := setupGin("POST", "/interface", controller.AddInterface)

	req, _ := http.NewRequest("POST", "/interface", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetClientArchive_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUsecaseService(ctrl)

	mockSvc.EXPECT().
		GetClientArchive().
		Return([]usecases.ClientResponse{
			{Public: "key1"},
		}, nil)

	controller := NewController(mockSvc, &config.ServerConfig{})

	r, w := setupGin("GET", "/clients/archive", controller.GetClientArchive)

	req, _ := http.NewRequest("GET", "/clients/archive", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetClientArchive_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUsecaseService(ctrl)

	mockSvc.EXPECT().
		GetClientArchive().
		Return(nil, errors.New("archive error"))

	controller := NewController(mockSvc, &config.ServerConfig{})

	r, w := setupGin("GET", "/clients/archive", controller.GetClientArchive)

	req, _ := http.NewRequest("GET", "/clients/archive", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetVersion_OK(t *testing.T) {
	controller := NewController(nil, &config.ServerConfig{})

	r, w := setupGin("GET", "/version", controller.GetVersion)

	req, _ := http.NewRequest("GET", "/version", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
