package wg

import (
	"errors"
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl"
)

type realWGFactory struct{}

func NewWGFactory() WGFactory {
	return &realWGFactory{}
}

func (f *realWGFactory) New() (WGClient, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	return client, nil // wgctrl.Client уже реализует нужные методы
}

func CheckUpInterface(factory WGFactory, ifname string) error {
	client, err := factory.New()
	if err != nil {
		return err
	}
	defer client.Close()

	devices, err := client.Devices()
	if err != nil {
		return err
	}

	for _, device := range devices {
		if device.Name == ifname {
			return errors.New(fmt.Sprintf("exist up interface %s", ifname))
		}
	}
	return nil
}
