package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"wireguard_api/config"
	"wireguard_api/controllers"

	"wireguard_api/usecases"

	"github.com/gin-gonic/gin"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func setupCtrlDeleteServer(t *testing.T) (
	*controllers.Controller,
	*MockService,
	func(),) {
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	cfg := &config.ServerConfig{
		DeleteInterface: true,
	}
	controller := controllers.NewController(
		&usecases.Usecases{Service: mockService},
		cfg,
	)
	cleanup := func() {
		ctrl.Finish()
	}
	return controller, mockService, cleanup
}

func TestCtrlDeleteServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		controller, mockService, cleanup := setupCtrlDeleteServer(t)
		defer cleanup()

		mockService.EXPECT().
			DeleteServer("priv", "wg0").
			Return(nil)

		body := `{"Private":"priv","Ifname":"wg0"}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(
			http.MethodPost,
			"/",
			bytes.NewBufferString(body),
		)
		c.Request.Header.Set("Content-Type", "application/json")

		controller.CtrlDeleteServer(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ok")
	})
}
