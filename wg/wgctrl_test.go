package wg

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestCheckUpInterface_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockWGFactory(ctrl)
	mockClient := NewMockWGClient(ctrl)

	mockFactory.EXPECT().
		New().
		Return(mockClient, nil)

	mockClient.EXPECT().
		Devices().
		Return([]*wgtypes.Device{
			{Name: "wg1"},
			{Name: "wg2"},
		}, nil)

	mockClient.EXPECT().
		Close().
		Return(nil)

	err := CheckUpInterface(mockFactory, "wg0")
	assert.NoError(t, err)
}

func TestCheckUpInterface_NewError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockWGFactory(ctrl)

	mockFactory.EXPECT().
		New().
		Return(nil, errors.New("wg error"))

	err := CheckUpInterface(mockFactory, "wg0")
	assert.Error(t, err)
}

func TestCheckUpInterface_DevicesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockWGFactory(ctrl)
	mockClient := NewMockWGClient(ctrl)

	mockFactory.EXPECT().
		New().
		Return(mockClient, nil)

	mockClient.EXPECT().
		Devices().
		Return(nil, errors.New("devices error"))

	mockClient.EXPECT().
		Close().
		Return(nil)

	err := CheckUpInterface(mockFactory, "wg0")
	assert.Error(t, err)
}

var _ WGFactory = (*realWGFactory)(nil)
