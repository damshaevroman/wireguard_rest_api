package usecases

// import (
// 	"context"
// 	"errors"
// 	"testing"
// 	"time"
// 	"wireguard_api/db"
// 	"wireguard_api/usecases"

// 	"github.com/golang/mock/gomock"
// 	"github.com/stretchr/testify/assert"
// )

// // Тестирование NewClient
// func TestNewClient_Errors(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockClientRepo := usecases.NewMockClientRepo(ctrl)
// 	mockServerRepo := usecases.NewMockServerRepo(ctrl)
// 	mockIpTables := usecases.NewMockIPTables(ctrl)
// 	mockPing := usecases.NewMockPingService(ctrl)

// 	u := &usecases.Usecases{
// 		ClientRepo: mockClientRepo,
// 		ServerRepo: mockServerRepo,
// 		IpTables:   mockIpTables,
// 		PingStatus: mockPing,
// 	}

// 	// Ошибка при checkIpMask
// 	err := u.checkIpMask("nonexistent", "10.0.0.1/24")
// 	assert.Error(t, err)

// 	// Ошибка при генерации IP
// 	mockClientRepo.EXPECT().GetPublicEnpointPort(gomock.Any()).Return(db.ServerCert{}, errors.New("repo error"))
// 	_, err = u.NewClient("wg0", "10.0.0.1/24", "10.0.0.0/24")
// 	assert.Error(t, err)
// }

// // Тестирование GetStatus
// func TestGetStatus(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	u := &usecases.Usecases{}
// 	// Положительный кейс (возврат пустого списка)
// 	status, err := u.GetStatus()
// 	assert.NoError(t, err)
// 	assert.NotNil(t, status)
// }

// // Тестирование GetAllClients
// func TestGetAllClients(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockClientRepo := usecases.NewMockClientRepo(ctrl)
// 	mockPing := usecases.NewMockPingService(ctrl)

// 	u := &usecases.Usecases{
// 		ClientRepo: mockClientRepo,
// 		PingStatus: mockPing,
// 	}

// 	mockClientRepo.EXPECT().GetAllClient().Return([]db.ClientCert{
// 		{Ifname: "wg0", IP: "10.0.0.1/24"},
// 	}, nil)

// 	clients, err := u.GetAllClients()
// 	assert.NoError(t, err)
// 	assert.Len(t, clients, 1)
// }

// // Тестирование DeleteClient
// func TestDeleteClient(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockClientRepo := usecases.NewMockClientRepo(ctrl)
// 	mockPing := usecases.NewMockPingService(ctrl)

// 	u := &usecases.Usecases{
// 		ClientRepo: mockClientRepo,
// 		PingStatus: mockPing,
// 	}

// 	mockClientRepo.EXPECT().DeleteClientCert("pub").Return(db.ClientCert{Ifname: "wg0", Public: "pub", IP: "10.0.0.1/24"}, nil)
// 	mockPing.EXPECT().Delete("10.0.0.1")

// 	err := u.DeleteClient("pub")
// 	assert.NoError(t, err)
// }

// // Тестирование GetClientArchive
// func TestGetClientArchive(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockClientRepo := usecases.NewMockClientRepo(ctrl)
// 	u := &usecases.Usecases{
// 		ClientRepo: mockClientRepo,
// 	}

// 	mockClientRepo.EXPECT().GetClientArchive().Return([]db.ClientCert{
// 		{Ifname: "wg0", IP: "10.0.0.1/24"},
// 	}, nil)

// 	data, err := u.GetClientArchive()
// 	assert.NoError(t, err)
// 	assert.Len(t, data, 1)
// }

// // Тестирование NewInterface с ошибкой создания интерфейса
// func TestNewInterface_Errors(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockServerRepo := usecases.NewMockServerRepo(ctrl)
// 	u := &usecases.Usecases{
// 		ServerRepo: mockServerRepo,
// 	}

// 	mockServerRepo.EXPECT().CreateServerCert(gomock.Any()).Return(errors.New("repo error"))

// 	_, err := u.NewInterface("wg0", "10.0.0.1/24", "endpoint", 51820)
// 	assert.Error(t, err)
// }

// // Тестирование DeleteServer
// func TestDeleteServer(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockServerRepo := usecases.NewMockServerRepo(ctrl)
// 	u := &usecases.Usecases{
// 		ServerRepo: mockServerRepo,
// 	}

// 	mockServerRepo.EXPECT().DeleteServer("priv", "wg0").Return(nil)
// 	err := u.DeleteServer("priv", "wg0")
// 	assert.Error(t, err) // будет ошибка, так как stopInterface вызывает exec.Command (можно замокать для интеграции)
// }

// // Тестирование SetUsForward с ошибкой
// func TestSetUsForward_Errors(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockIpTables := usecases.NewMockIPTables(ctrl)
// 	mockServerRepo := usecases.NewMockServerRepo(ctrl)

// 	u := &usecases.Usecases{
// 		IpTables:   mockIpTables,
// 		ServerRepo: mockServerRepo,
// 	}

// 	mockIpTables.EXPECT().SetForward(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("iptables error"))
// 	mockServerRepo.EXPECT().DeleteForward("test").Return(nil)

// 	err := u.SetUsForward(1, "ACCEPT", "delete", "src", "dst", "tcp", "80", "test", false, false)
// 	assert.Error(t, err)
// }

// // Тестирование SetUsMasquerade
// func TestSetUsMasquerade_Errors(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockIpTables := usecases.NewMockIPTables(ctrl)
// 	mockServerRepo := usecases.NewMockServerRepo(ctrl)

// 	u := &usecases.Usecases{
// 		IpTables:   mockIpTables,
// 		ServerRepo: mockServerRepo,
// 	}

// 	mockIpTables.EXPECT().SetMasquerade("write", "src", "ifname", "comment").Return(errors.New("iptable error"))
// 	err := u.SetUsMasquerade("write", "src", "ifname", "comment")
// 	assert.Error(t, err)
// }

// // Тестирование GetServerArchive
// func TestGetServerArchive(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
// 	mockServerRepo := usecases.NewMockServerRepo(ctrl)

// 	u := &usecases.Usecases{
// 		ServerRepo: mockServerRepo,
// 	}

// 	mockServerRepo.EXPECT().GetServerArchive().Return([]db.ServerCert{
// 		{Ifname: "wg0", IP: "10.0.0.1"},
// 	}, nil)

// 	data, err := u.GetServerArchive()
// 	assert.NoError(t, err)
// 	assert.Len(t, data, 1)
// }

// // Тестирование GetIptablesRules
// func TestGetIptablesRules(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
// 	mockServerRepo := usecases.NewMockServerRepo(ctrl)
// 	mockIpTables := usecases.NewMockIPTables(ctrl)

// 	u := &usecases.Usecases{
// 		ServerRepo: mockServerRepo,
// 		IpTables:   mockIpTables,
// 	}

// 	mockServerRepo.EXPECT().GetForward().Return([]db.Forward{}, nil)
// 	mockServerRepo.EXPECT().GetMasquerade().Return([]db.Masquerade{}, nil)

// 	data, err := u.GetIptablesRules()
// 	assert.NoError(t, err)
// 	assert.NotNil(t, data)
// }

// // Тест PingLoop с контекстом
// func TestPingLoop_ContextDone(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockClientRepo := usecases.NewMockClientRepo(ctrl)
// 	mockPing := usecases.NewMockPingService(ctrl)
// 	u := &usecases.Usecases{
// 		ClientRepo: mockClientRepo,
// 		PingStatus: mockPing,
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())
// 	cancel() // сразу закрываем

// 	go u.PingLoop(ctx)
// 	time.Sleep(10 * time.Millisecond)
// }
