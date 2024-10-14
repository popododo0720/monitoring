package process

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"
)

type ProcessInfo struct {
	User    string
	PID     string
	CPU     float64
	MEM     float64
	VSZ     string
	RSS     string
	TTY     string
	STAT    string
	START   string
	TIME    string
	COMMAND string
}

var processes []ProcessInfo

func GetProcessList() ([]ProcessInfo, error) {
	processes = []ProcessInfo{}

	cmd := exec.Command("sh", "-c", "ps aux | awk '!/grep/ && !/ps aux/'")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	if scanner.Scan() {
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 11 {
			continue
		}

		command := strings.Join(fields[10:], " ")

		cpu, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			cpu = 0
		}

		mem, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			mem = 0
		}

		process := ProcessInfo{
			User:    fields[0],
			PID:     fields[1],
			CPU:     cpu,
			MEM:     mem,
			VSZ:     fields[4],
			RSS:     fields[5],
			TTY:     fields[6],
			STAT:    fields[7],
			START:   fields[8],
			TIME:    fields[9],
			COMMAND: command, // COMMAND 필드는 공백 포함
		}

		// 프로세스 리스트에 추가
		processes = append(processes, process)
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return processes, nil
}
