package controllers

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"wireguard_api/config"
	"wireguard_api/usecases"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCtrlStopServer_OK(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		StopInterface("wg0").
		Return(nil)

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	body := `{"ifname":"wg0"}`
	r, w := setupGin("POST", "/stop", ctrl.CtrlStopServer)

	req, _ := http.NewRequest("POST", "/stop", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCtrlStopServer_Error(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		StopInterface(gomock.Any()).
		Return(errors.New("stop error"))

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	body := `{"ifname":"wg0"}`
	r, w := setupGin("POST", "/stop", ctrl.CtrlStopServer)

	req, _ := http.NewRequest("POST", "/stop", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCtrlDeleteServer_OK(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		DeleteServer("priv", "wg0").
		Return(nil)

	cfg := &config.ServerConfig{DeleteInterface: true}
	ctrl := NewController(mockSvc, cfg)

	body := `{"private":"priv","ifname":"wg0"}`
	r, w := setupGin("DELETE", "/server", ctrl.CtrlDeleteServer)

	req, _ := http.NewRequest("DELETE", "/server", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCtrlDeleteServer_Error(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		DeleteServer(gomock.Any(), gomock.Any()).
		Return(errors.New("delete error"))

	cfg := &config.ServerConfig{DeleteInterface: true}
	ctrl := NewController(mockSvc, cfg)

	body := `{"private":"priv","ifname":"wg0"}`
	r, w := setupGin("DELETE", "/server", ctrl.CtrlDeleteServer)

	req, _ := http.NewRequest("DELETE", "/server", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCtrlStartServer_OK(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		StartInterface("wg0").
		Return(nil)

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	body := `{"ifname":"wg0"}`
	r, w := setupGin("POST", "/start", ctrl.CtrlStartServer)

	req, _ := http.NewRequest("POST", "/start", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCtrlStartServer_Error(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		StartInterface(gomock.Any()).
		Return(errors.New("start error"))

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	body := `{"ifname":"wg0"}`
	r, w := setupGin("POST", "/start", ctrl.CtrlStartServer)

	req, _ := http.NewRequest("POST", "/start", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSetForward_OK(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		SetUsForward(
			gomock.Any(),        // position
			gomock.Eq("ACCEPT"), // action
			gomock.Eq("add"),
			gomock.Eq("0.0.0.0/0"),
			gomock.Eq("10.0.0.0/24"),
			gomock.Eq("tcp"),
			gomock.Eq("80"),
			gomock.Eq("test_comment"),
			gomock.Eq(true),
			gomock.Eq(false),
		).
		Return(nil)
	ctrl := NewController(mockSvc, &config.ServerConfig{})

	body := `{
		"position":1,
		"action":"ACCEPT",
		"command":"add",
		"source":"0.0.0.0/0",
		"destination":"10.0.0.0/24",
		"protocol":"tcp",
		"port":"80",
		"comment":"test_comment",
		"list":true,
		"except":false
	}`

	r, w := setupGin("POST", "/forward", ctrl.SetForward)
	req, _ := http.NewRequest("POST", "/forward", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetForward_Error(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		SetUsForward(
			gomock.Any(),        // position
			gomock.Eq("ACCEPT"), // action
			gomock.Eq("add"),
			gomock.Eq("0.0.0.0/0"),
			gomock.Eq("10.0.0.0/24"),
			gomock.Eq("tcp"),
			gomock.Eq("80"),
			gomock.Eq("test_comment"),
			gomock.Eq(true),
			gomock.Eq(false),
		).Return(errors.New("update error"))

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	body := `{
		"position":1,
		"action":"ACCEPT",
		"command":"add",
		"source":"0.0.0.0/0",
		"destination":"10.0.0.0/24",
		"protocol":"tcp",
		"port":"80",
		"comment":"test_comment",
		"list":true,
		"except":false
	}`
	r, w := setupGin("POST", "/forward", ctrl.SetForward)

	req, _ := http.NewRequest("POST", "/forward", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSetForwardUpdateList_OK(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		UpdateIpSetList(
			"add",
			"test_set",
			[]string{"1.1.1.1", "2.2.2.2"},
			false,
		).
		Return(nil)

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	body := `{
		"command":"add",
		"ipset_name":"test_set",
		"ip_list":["1.1.1.1","2.2.2.2"],
		"single":false
	}`

	r, w := setupGin("POST", "/forward/list", ctrl.SetForwardUpdateList)
	req, _ := http.NewRequest("POST", "/forward/list", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetForwardUpdateList_Error(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		UpdateIpSetList(
			gomock.Eq("add"),
			gomock.Eq("snt"),
			gomock.Eq([]string{"10.180.180.43"}),
			gomock.Eq(false),
		).
		Return(errors.New("update error"))

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	body := `{
		"command": "add",
		"ip_list": ["10.180.180.43"],
		"ipset_name": "snt",
		"single": false
	  }`
	r, w := setupGin("POST", "/forward/list", ctrl.SetForwardUpdateList)

	req, _ := http.NewRequest("POST", "/forward/list", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSetMasquerade_OK(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		SetUsMasquerade(
			"add",
			"0.0.0.0/0",
			"eth0",
			"test_comment",
		).
		Return(nil)

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	body := `{
		"command":"add",
		"source":"0.0.0.0/0",
		"ifname":"eth0",
		"comment":"test comment"
	}`

	r, w := setupGin("POST", "/masquerade", ctrl.SetMasquerade)
	req, _ := http.NewRequest("POST", "/masquerade", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetMasquerade_Error(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		SetUsMasquerade(
			"add",
			"0.0.0.0/0",
			"eth0",
			"test_comment",
		).
		Return(errors.New("masquerade error"))

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	body := `{
		"command":"add",
		"source":"0.0.0.0/0",
		"ifname":"eth0",
		"comment":"test comment"
	}`

	r, w := setupGin("POST", "/masquerade", ctrl.SetMasquerade)

	req, _ := http.NewRequest("POST", "/masquerade", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestCtrlGetServerArchive_OK(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		GetServerArchive().
		Return([]usecases.ServerInterfaces{
			{Ifname: "wg0"},
		}, nil)

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	r, w := setupGin("GET", "/server/archive", ctrl.CtrlGetServerArchive)
	req, _ := http.NewRequest("GET", "/server/archive", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCtrlGetInterfaces_OK(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		GetServerInterfaces().
		Return([]usecases.ServerInterfaces{
			{Ifname: "wg0"},
		}, nil)

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	r, w := setupGin("GET", "/server/interfaces", ctrl.CtrlGetInterfaces)
	req, _ := http.NewRequest("GET", "/server/interfaces", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCtrlGetInterfaces_Error(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		GetServerInterfaces().
		Return(nil, errors.New("interfaces error"))

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	r, w := setupGin("GET", "/server/interfaces", ctrl.CtrlGetInterfaces)
	req, _ := http.NewRequest("GET", "/server/interfaces", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCtrlGetIptables_OK(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		GetIptablesRules().
		Return(usecases.IptablesRulesData{}, nil)

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	r, w := setupGin("GET", "/iptables", ctrl.CtrlGetIptables)
	req, _ := http.NewRequest("GET", "/iptables", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCtrlGetIptables_Error(t *testing.T) {
	gc := gomock.NewController(t)
	defer gc.Finish()

	mockSvc := NewMockUsecaseService(gc)
	mockSvc.EXPECT().
		GetIptablesRules().
		Return(usecases.IptablesRulesData{}, errors.New("iptables error"))

	ctrl := NewController(mockSvc, &config.ServerConfig{})

	r, w := setupGin("GET", "/iptables", ctrl.CtrlGetIptables)
	req, _ := http.NewRequest("GET", "/iptables", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
