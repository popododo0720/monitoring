package instance_metrics

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// 디스크 총량
func GetInstanceDiskSize() (float64, error) {
	cmd := exec.Command("sh", "-c", "df -m | awk '$6 == \"/\" {print $2}'")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("error executing command: %w", err)
	}

	output := strings.TrimSpace(out.String())

	diskSize, _ := strconv.ParseFloat(output, 64)

	return diskSize, nil
}

// 디스크사용량
func GetInstanceDiskUsed() (float64, error) {
	cmd := exec.Command("sh", "-c", "df -m | awk '$6 == \"/\" {print $3}'")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("error executing command: %w", err)
	}

	output := strings.TrimSpace(out.String())

	diskUsed, _ := strconv.ParseFloat(output, 64)

	return diskUsed, nil
}

// 디스크 남은용량
func GetInstanceDiskAvail() (float64, error) {
	cmd := exec.Command("sh", "-c", "df -m | awk '$6 == \"/\" {print $4}'")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("error executing command: %w", err)
	}

	output := strings.TrimSpace(out.String())

	diskAvail, _ := strconv.ParseFloat(output, 64)

	return diskAvail, nil
}

// 퍼센트
func GetInstanceDiskUseRate() (float64, error) {
	cmd := exec.Command("sh", "-c", "df -m | awk '$6 == \"/\" {print $5}'")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("error executing command: %w", err)
	}

	output := strings.TrimSpace(out.String())

	output = strings.TrimSuffix(output, "%")

	diskUseRate, err := strconv.ParseFloat(output, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting output to float: %w", err)
	}

	return diskUseRate, nil
}
