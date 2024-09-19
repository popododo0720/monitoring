package instance_metrics

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func GetInstanceMemUsage() (float64, error) {
	totalCmd := exec.Command("sh", "-c", "free -m | grep Mem | awk '{print $2}'")
	var totalOut bytes.Buffer
	totalCmd.Stdout = &totalOut
	err := totalCmd.Run()
	if err != nil {
		return 0, fmt.Errorf("error executing command to get total memory: %w", err)
	}

	totalOutput := strings.TrimSpace(totalOut.String())
	totalMemory, _ := strconv.Atoi(totalOutput)

	usedCmd := exec.Command("sh", "-c", "free -m | grep Mem | awk '{print $3}'")
	var usedOut bytes.Buffer
	usedCmd.Stdout = &usedOut
	err = usedCmd.Run()
	if err != nil {
		return 0, fmt.Errorf("error executing command to get used memory: %w", err)
	}

	usedOutput := strings.TrimSpace(usedOut.String())
	usedMemory, _ := strconv.Atoi(usedOutput)

	memUsage := (float64(usedMemory) / float64(totalMemory)) * 100

	memUsageFormatted := fmt.Sprintf("%.1f", memUsage)

	memUsageFloat, _ := strconv.ParseFloat(memUsageFormatted, 64)

	return memUsageFloat, nil
}
