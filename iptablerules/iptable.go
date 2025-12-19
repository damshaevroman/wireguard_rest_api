package iptablerules

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/coreos/go-iptables/iptables"
)

func NewIptables(table IptablesInterface, runner CommandRunner) *IptablesStruct {
	return &IptablesStruct{
		table:  table,
		runner: runner,
	}
}
func CreateGoIptables() (*iptables.IPTables, error) {
	t, err := iptables.New()
	if err != nil {
		return nil, fmt.Errorf("iptables init failed: %w", err)
	}
	return t, nil
}

func Init(table IptablesInterface) *IptablesStruct {
	return &IptablesStruct{
		table: table,
	}
}

func (i *IptablesStruct) checkTypePiort(typePort string) bool {
	switch typePort {
	case "tcp":
		return true
	case "udp":
		return true
	case "icmp":
		return true
	default:
		return false
	}
}

func (i *IptablesStruct) SetForwardList(position int, port, action, command, source, destination, protocol, comment string, except bool) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if !i.checkTypePiort(protocol) {
		return errors.New("typePort can be: tcp, udp, icmp")
	}

	posStr := strconv.Itoa(position)
	stringPort := strings.TrimSpace(port)
	icmpComment := "icmp_" + comment

	buildICMPArgs := func(cmd string) []string {
		args := []string{}

		switch cmd {
		case "write":
			args = append(args, "-I", "FORWARD", posStr)
		case "delete":
			args = append(args, "-D", "FORWARD")
		}

		args = append(args, "-s", source)
		args = append(args, "-m", "set")
		if !except {
			args = append(args, "!")
		}

		args = append(args, "--match-set", destination, "dst")
		args = append(args,
			"-p", "icmp",
			"-j", action,
			"-m", "comment",
			"--comment", icmpComment,
		)

		return args
	}

	buildPortArgs := func() []string {
		args := []string{"-s", source, "-m", "set"}
		if !except {
			args = append(args, "!")
		}

		args = append(args, "--match-set", destination, "dst")

		if stringPort != "" {
			args = append(args, "-p", protocol, "-m", "multiport", "--dport", stringPort)
		}

		args = append(args,
			"-j", action,
			"-m", "comment",
			"--comment", comment,
		)

		return args
	}

	if command == "write" {
		out, err := i.runner.Run("iptables", "-nvL")
		if err != nil {
			log.Printf("SetForwardList: %s", err.Error())
			return errors.New(string(out))
		}

		if !strings.Contains(string(out), icmpComment) {
			args := buildICMPArgs("write")
			out, err := i.runner.Run("iptables", args...)
			if err != nil {
				log.Printf("SetForwardList: %s", err.Error())
				return errors.New(string(out))
			}
		}
	}

	args := buildPortArgs()
	switch command {
	case "write":
		if err := i.table.InsertUnique("filter", "FORWARD", position, args...); err != nil {
			return fmt.Errorf("SetForwardList: %v", err)
		}
	case "delete":
		icmpArgs := buildICMPArgs("delete")
		out, err := i.runner.Run("iptables", icmpArgs...)
		if err != nil {
			log.Printf("SetForwardList delete icmp: %s", err.Error())
			return errors.New(string(out))
		}
		if err := i.table.DeleteIfExists("filter", "FORWARD", args...); err != nil {
			return fmt.Errorf("SetForwardList delete: %v", err)
		}
	default:
		if command == "" {
			return errors.New("empty value of command")
		}
		return fmt.Errorf("command not found: %s", command)
	}

	return nil
}

func (i *IptablesStruct) ruleExists(source, destination, comment string) error {
	out, err := i.runner.Run("iptables-save")
	if err != nil {
		return fmt.Errorf("cannot read iptables-save: %w", err)
	}

	ruleSignature := fmt.Sprintf("-A FORWARD -s %s -d %s -p icmp -m comment --comment %q -j ACCEPT", source, destination, comment)
	if bytes.Contains(out, []byte(ruleSignature)) {
		return fmt.Errorf("this rule already exist")
	}
	return nil
}

func (i *IptablesStruct) SetForward(position int, port, action, command, source, destination, protocol, comment string, except bool) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	port = strings.TrimSpace(port)
	if !i.checkTypePiort(protocol) {
		return errors.New("typePort can be: tcp, udp, icmp")
	}

	posStr := strconv.Itoa(position)
	args := []string{}
	switch command {
	case "write":
		if protocol == "icmp" {
			if err := i.ruleExists(source, destination, comment); err != nil {
				return err
			}
			args = append(args, "-I", "FORWARD", posStr, "-s", source)
			if !except {
				args = append(args, "!")
			}
			args = append(args, "-d", destination, "-p", "icmp", "-j", action, "-m", "comment", "--comment", comment)

			out, err := i.runner.Run("iptables", args...)
			if err != nil {
				log.Printf("Error SetForward - write: %s\n%s", err.Error(), string(out))
				return err
			}
			return nil
		}

		args = []string{"-s", source}
		if !except {
			args = append(args, "!")
		}
		args = append(args, "-d", destination, "-j", action, "-m", "comment", "--comment", comment)

		if port != "" {
			args = append(args, "-p", protocol, "-m", "multiport", "--dport", port)
		}

		if err := i.table.InsertUnique("filter", "FORWARD", position, args...); err != nil {
			return fmt.Errorf("add error: %v", err)
		}
		return nil

	case "delete":
		args = []string{"-s", source}
		if !except {
			args = append(args, "!")
		}
		args = append(args, "-d", destination, "-j", action, "-m", "comment", "--comment", comment)

		if protocol == "icmp" {
			args = append(args, "-p", "icmp")
		} else if port != "" {
			args = append(args, "-p", protocol, "-m", "multiport", "--dport", port)
		}

		if err := i.table.DeleteIfExists("filter", "FORWARD", args...); err != nil {
			return fmt.Errorf("delete error: %v", err)
		}
		return nil
	}

	if command == "" {
		return errors.New("empty value of command")
	}
	return fmt.Errorf("command not found: %s", command)
}

func (i *IptablesStruct) SetMasquerade(command, subnet, ifname, comment string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	switch command {
	case "write":
		err := i.table.InsertUnique("nat", "POSTROUTING", 1, "-s", subnet, "-o", ifname, "-j", "MASQUERADE", "-m", "comment", "--comment", comment)
		if err != nil {
			return err
		}
		return nil
	case "delete":
		err := i.table.DeleteIfExists("nat", "POSTROUTING", "-s", subnet, "-o", ifname, "-j", "MASQUERADE", "-m", "comment", "--comment", comment)
		if err != nil {
			return err
		}
		return nil

	default:
		if command == "" {
			return errors.New("iptable command value is empty")
		}
		return fmt.Errorf("iptable did not find command %s", command)
	}
}

func (i *IptablesStruct) GetMasqueradeList() ([]string, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	masqList, err := i.table.List("nat", "POSTROUTING")
	if err != nil {
		return []string{}, err
	}
	return masqList, nil
}

func (i *IptablesStruct) GetForwardList() ([]string, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	frwdList, err := i.table.List("filter", "FORWARD")
	if err != nil {
		return []string{}, err
	}

	return frwdList, nil
}

func (i *IptablesStruct) FlushForward() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	err := i.table.ClearChain("filter", "FORWARD")
	if err != nil {
		return err
	}

	return nil
}
