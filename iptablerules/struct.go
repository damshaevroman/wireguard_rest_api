package iptablerules

import (
	"os/exec"
	"sync"
)

type RealIptablesRunner struct{}

func (r *RealIptablesRunner) Run(args ...string) ([]byte, error) {
	return exec.Command("iptables", args...).CombinedOutput()
}

type IptablesStruct struct {
	mu     sync.Mutex
	table  IptablesInterface
	runner CommandRunner
}
