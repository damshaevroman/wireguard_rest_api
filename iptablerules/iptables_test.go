package iptablerules

import (
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCheckTypePort(t *testing.T) {
	ipt := &IptablesStruct{}

	tests := []struct {
		proto string
		want  bool
	}{
		{"tcp", true},
		{"udp", true},
		{"icmp", true},
		{"http", false},
	}

	for _, tt := range tests {
		if got := ipt.checkTypePiort(tt.proto); got != tt.want {
			t.Fatalf("proto %s expected %v, got %v", tt.proto, tt.want, got)
		}
	}
}

func TestSetForwardList_Write_TCP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRunner := NewMockCommandRunner(ctrl)
	mockTable := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		table:  mockTable,
		runner: mockRunner,
	}

	// 1. iptables -nvL (проверка icmp правила)
	mockRunner.EXPECT().
		Run("iptables", "-nvL").
		Return([]byte(""), nil)

	// 2. Добавление ICMP правила (если не найдено)
	mockRunner.EXPECT().
		Run(
			"iptables",
			"-I", "FORWARD", "1",
			"-s", "192.168.1.1",
			"-m", "set",
			"!",
			"--match-set", "192.168.1.0/24", "dst",
			"-p", "icmp",
			"-j", "ACCEPT",
			"-m", "comment",
			"--comment", "icmp_test-comment",
		).
		Return([]byte(""), nil)

	// 3. Добавление TCP правила через InsertUnique
	mockTable.EXPECT().
		InsertUnique(
			"filter",
			"FORWARD",
			1,
			"-s", "192.168.1.1",
			"-m", "set",
			"!",
			"--match-set", "192.168.1.0/24", "dst",
			"-p", "tcp",
			"-m", "multiport",
			"--dport", "80",
			"-j", "ACCEPT",
			"-m", "comment",
			"--comment", "test-comment",
		).
		Return(nil)

	err := ipt.SetForwardList(
		1,
		"80",
		"ACCEPT",
		"write",
		"192.168.1.1",
		"192.168.1.0/24",
		"tcp",
		"test-comment",
		false,
	)

	assert.NoError(t, err)
}

func TestSetForwardList_Write_TCP_IptablesListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner := NewMockCommandRunner(ctrl)
	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		runner: runner,
		table:  table,
	}

	runner.EXPECT().
		Run("iptables", "-nvL").
		Return([]byte("permission denied"), errors.New("exit 1"))

	err := ipt.SetForwardList(
		1, "80", "ACCEPT", "write",
		"192.168.1.1", "192.168.1.0/24",
		"tcp", "test-comment", false,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestSetForwardList_Write_TCP_ICMPInsertError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner := NewMockCommandRunner(ctrl)
	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		runner: runner,
		table:  table,
	}

	gomock.InOrder(
		runner.EXPECT().
			Run("iptables", "-nvL").
			Return([]byte(""), nil),

		runner.EXPECT().
			Run(gomock.Any(), gomock.Any()).
			Return([]byte("icmp failed"), errors.New("exit 1")),
	)

	err := ipt.SetForwardList(
		1, "80", "ACCEPT", "write",
		"192.168.1.1", "192.168.1.0/24",
		"tcp", "test-comment", false,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "icmp failed")
}

func TestSetForwardList_Write_TCP_InsertUniqueError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner := NewMockCommandRunner(ctrl)
	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		runner: runner,
		table:  table,
	}

	gomock.InOrder(
		runner.EXPECT().
			Run("iptables", "-nvL").
			Return([]byte(""), nil),

		runner.EXPECT().
			Run(gomock.Any(), gomock.Any()).
			Return([]byte(""), nil),
	)

	table.EXPECT().
		InsertUnique(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("insert failed"))

	err := ipt.SetForwardList(
		1, "80", "ACCEPT", "write",
		"192.168.1.1", "192.168.1.0/24",
		"tcp", "test-comment", false,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SetForwardList")
}

func TestSetForwardList_InvalidCommand(t *testing.T) {
	ipt := &IptablesStruct{}

	err := ipt.SetForwardList(
		1, "80", "ACCEPT", "unknown",
		"192.168.1.1", "192.168.1.0/24",
		"tcp", "test", false,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command not found")
}

func TestRuleExists_Exists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner := NewMockCommandRunner(ctrl)

	ipt := &IptablesStruct{
		runner: runner,
	}

	runner.EXPECT().
		Run("iptables-save").
		Return([]byte(
			`-A FORWARD -s 1.1.1.1 -d 2.2.2.2 -p icmp -m comment --comment "test" -j ACCEPT`,
		), nil)

	err := ipt.ruleExists("1.1.1.1", "2.2.2.2", "test")
	assert.Error(t, err)
}

func TestRuleExists_Exists_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner := NewMockCommandRunner(ctrl)

	ipt := &IptablesStruct{
		runner: runner,
	}

	runner.EXPECT().
		Run("iptables-save").
		Return([]byte(
			`-A FORWARD -s 1.1.1.1 -d 2.2.2.2 -p icmp -m comment --comment "test" -j ACCEPT`,
		), errors.New("failed"))

	err := ipt.ruleExists("1.1.1.1", "2.2.2.2", "test")
	assert.Error(t, err)
	assert.NotNil(t, err)
}

func TestRuleExists_NotExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner := NewMockCommandRunner(ctrl)

	ipt := &IptablesStruct{
		runner: runner,
	}

	runner.EXPECT().
		Run("iptables-save").
		Return([]byte(""), nil)

	err := ipt.ruleExists("1.1.1.1", "2.2.2.2", "test")
	assert.NoError(t, err)
}

func TestRuleExists_NotExists_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner := NewMockCommandRunner(ctrl)

	ipt := &IptablesStruct{
		runner: runner,
	}

	runner.EXPECT().
		Run("iptables-save").
		Return([]byte(""), errors.New("failed"))

	err := ipt.ruleExists("1.1.1.1", "2.2.2.2", "test")
	assert.Error(t, err)

}

func TestSetForward_Write_ICMP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner := NewMockCommandRunner(ctrl)
	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		runner: runner,
		table:  table,
	}

	gomock.InOrder(
		runner.EXPECT().
			Run("iptables-save").
			Return([]byte(""), nil),

		runner.EXPECT().
			Run(
				"iptables",
				"-I", "FORWARD", "1",
				"-s", "1.1.1.1",
				"!",
				"-d", "2.2.2.2",
				"-p", "icmp",
				"-j", "ACCEPT",
				"-m", "comment",
				"--comment", "test",
			).
			Return([]byte(""), nil),
	)

	err := ipt.SetForward(
		1, "", "ACCEPT", "write",
		"1.1.1.1", "2.2.2.2",
		"icmp", "test", false,
	)

	assert.NoError(t, err)
}

func TestSetForward_Write_ICMP_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner := NewMockCommandRunner(ctrl)
	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		runner: runner,
		table:  table,
	}

	gomock.InOrder(
		runner.EXPECT().
			Run("iptables-save").
			Return([]byte(""), nil),

		runner.EXPECT().
			Run(
				"iptables",
				"-I", "FORWARD", "1",
				"-s", "1.1.1.1",
				"!",
				"-d", "2.2.2.2",
				"-p", "icmp",
				"-j", "ACCEPT",
				"-m", "comment",
				"--comment", "test",
			).
			Return([]byte(""), errors.New("list error")),
	)

	err := ipt.SetForward(
		1, "", "ACCEPT", "write",
		"1.1.1.1", "2.2.2.2",
		"icmp", "test", false,
	)

	assert.Error(t, err)
}

func TestSetForward_Write_TCP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		table: table,
	}

	table.EXPECT().
		InsertUnique(
			"filter", "FORWARD", 1,
			"-s", "1.1.1.1",
			"!",
			"-d", "2.2.2.2",
			"-j", "ACCEPT",
			"-m", "comment",
			"--comment", "test",
			"-p", "tcp",
			"-m", "multiport",
			"--dport", "80",
		).
		Return(nil)

	err := ipt.SetForward(
		1, "80", "ACCEPT", "write",
		"1.1.1.1", "2.2.2.2",
		"tcp", "test", false,
	)

	assert.NoError(t, err)
}

func TestSetForward_Write_TCP_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		table: table,
	}
	expectedErr := errors.New("list error")
	table.EXPECT().
		InsertUnique(
			"filter", "FORWARD", 1,
			"-s", "1.1.1.1",
			"!",
			"-d", "2.2.2.2",
			"-j", "ACCEPT",
			"-m", "comment",
			"--comment", "test",
			"-p", "tcp",
			"-m", "multiport",
			"--dport", "80",
		).
		Return(expectedErr)

	err := ipt.SetForward(
		1, "80", "ACCEPT", "write",
		"1.1.1.1", "2.2.2.2",
		"tcp", "test", false,
	)

	assert.Error(t, err)
}

func TestSetForward_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		table: table,
	}

	table.EXPECT().
		DeleteIfExists(
			"filter", "FORWARD",
			"-s", "1.1.1.1",
			"!",
			"-d", "2.2.2.2",
			"-j", "ACCEPT",
			"-m", "comment",
			"--comment", "test",
			"-p", "tcp",
			"-m", "multiport",
			"--dport", "80",
		).
		Return(nil)

	err := ipt.SetForward(
		1, "80", "ACCEPT", "delete",
		"1.1.1.1", "2.2.2.2",
		"tcp", "test", false,
	)

	assert.NoError(t, err)
}

func TestSetForward_Delete_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		table: table,
	}

	table.EXPECT().
		DeleteIfExists(
			"filter", "FORWARD",
			"-s", "1.1.1.1",
			"!",
			"-d", "2.2.2.2",
			"-j", "ACCEPT",
			"-m", "comment",
			"--comment", "test",
			"-p", "tcp",
			"-m", "multiport",
			"--dport", "80",
		).
		Return(errors.New("delete failed"))

	err := ipt.SetForward(
		1, "80", "ACCEPT", "delete",
		"1.1.1.1", "2.2.2.2",
		"tcp", "test", false,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestSetMasquerade_Write(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		table: table,
	}

	table.EXPECT().
		InsertUnique(
			"nat", "POSTROUTING", 1,
			"-s", "10.0.0.0/24",
			"-o", "eth0",
			"-j", "MASQUERADE",
			"-m", "comment",
			"--comment", "test",
		).
		Return(nil)

	err := ipt.SetMasquerade("write", "10.0.0.0/24", "eth0", "test")
	assert.NoError(t, err)
}

func TestSetMasquerade_Write_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		table: table,
	}
	expectedErr := errors.New("list error")
	table.EXPECT().
		InsertUnique(
			"nat", "POSTROUTING", 1,
			"-s", "10.0.0.0/24",
			"-o", "eth0",
			"-j", "MASQUERADE",
			"-m", "comment",
			"--comment", "test",
		).
		Return(expectedErr)

	err := ipt.SetMasquerade("write", "10.0.0.0/24", "eth0", "test")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

}

func TestGetMasqueradeList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		table: table,
	}

	table.EXPECT().
		List("nat", "POSTROUTING").
		Return([]string{"rule1", "rule2"}, nil)

	list, err := ipt.GetMasqueradeList()
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestGetMasqueradeList_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	table := NewMockIptablesInterface(ctrl)
	expectedErr := errors.New("list error")

	table.EXPECT().
		List("nat", "POSTROUTING").
		Return(nil, expectedErr)

	ipt := &IptablesStruct{
		table: table,
	}

	list, err := ipt.GetMasqueradeList()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, list)
}

func TestFlushForward(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	table := NewMockIptablesInterface(ctrl)

	ipt := &IptablesStruct{
		table: table,
	}

	table.EXPECT().
		ClearChain("filter", "FORWARD").
		Return(nil)

	err := ipt.FlushForward()
	assert.NoError(t, err)
}

func TestFlushForward_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTable := NewMockIptablesInterface(ctrl)
	expectedErr := errors.New("clear error")

	mockTable.EXPECT().
		ClearChain("filter", "FORWARD").
		Return(expectedErr)

	ipt := &IptablesStruct{
		table: mockTable,
	}

	err := ipt.FlushForward()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
