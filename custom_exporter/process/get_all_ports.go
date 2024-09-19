package process

import (
	"bufio"
	"os/exec"
	"strings"
)

type PortInfo struct {
	State   string
	RecvQ   string
	SendQ   string
	Local   string
	Peer    string
	Process string
}

func GetPortList() ([]PortInfo, error) {
	cmd := exec.Command("ss", "-tnlp")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	var ports []PortInfo

	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		process := ""
		if len(fields) > 5 {
			process = fields[5]
		}

		port := PortInfo{
			State:   fields[0],
			RecvQ:   fields[1],
			SendQ:   fields[2],
			Local:   fields[3],
			Peer:    fields[4],
			Process: process,
		}

		ports = append(ports, port)
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return ports, nil
}
