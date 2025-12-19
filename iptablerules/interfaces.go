package iptablerules

type CommandRunner interface {
	Run(cmd string, args ...string) ([]byte, error)
}

type IptablesInterface interface {
	InsertUnique(table, chain string, pos int, rulespec ...string) error
	DeleteIfExists(table, chain string, rulespec ...string) error
	List(table, chain string) ([]string, error)
	ClearChain(table, chain string) error
}

type IptablesManager interface {
	SetForwardList(position int, port, action, command, source, destination, protocol, comment string, except bool) error
	SetForward(position int, port, action, command, source, destination, protocol, comment string, except bool) error

	SetMasquerade(command, subnet, ifname, comment string) error

	GetForwardList() ([]string, error)
	GetMasqueradeList() ([]string, error)

	FlushForward() error
}
