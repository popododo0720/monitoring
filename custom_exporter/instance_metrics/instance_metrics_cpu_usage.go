package instance_metrics

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func GetInstanceCpuUsage() (float64, error) {
	cmd := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | awk '{print $2}'")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("error executing command: %w", err)
	}

	output := strings.TrimSpace(out.String())

	cpuUsage, err := strconv.ParseFloat(output, 64)

	if err != nil {
		return 0, fmt.Errorf("error converting output to float: %w", err)
	}

	return cpuUsage, nil

}
