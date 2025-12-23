package wg

import "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

type WGClient interface {
	Devices() ([]*wgtypes.Device, error)
	Close() error
}

type WGFactory interface {
	New() (WGClient, error)
}
