package pingstatus

import (
	net "net"
	"sync"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func TestPing_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := NewMockICMPConn(ctrl)
	mockFactory := NewMockICMPFactory(ctrl)

	// factory
	mockFactory.EXPECT().
		Listen().
		Return(mockConn, nil)

	// send
	mockConn.EXPECT().
		WriteTo(gomock.Any(), gomock.Any()).
		Return(1, nil)

	// read
	mockConn.EXPECT().
		SetReadDeadline(gomock.Any()).
		Return(nil)

	mockConn.EXPECT().
		ReadFrom(gomock.Any()).
		DoAndReturn(func(buf []byte) (int, net.Addr, error) {
			msg := icmp.Message{
				Type: ipv4.ICMPTypeEchoReply,
				Code: 0,
				Body: &icmp.Echo{
					ID:   1234,
					Seq:  1,
					Data: []byte("HELLO-R-U-THERE"),
				},
			}
			data, _ := msg.Marshal(nil)
			copy(buf, data)
			return len(data), &net.IPAddr{}, nil
		})

	mockConn.EXPECT().
		Close().
		Return(nil)

	ps := Init(mockFactory)

	var wg sync.WaitGroup
	wg.Add(1)
	go ps.Ping("8.8.8.8", &wg)
	wg.Wait()

	ok, duration := ps.Read("8.8.8.8")
	assert.True(t, ok)
	assert.NotZero(t, duration)
}
